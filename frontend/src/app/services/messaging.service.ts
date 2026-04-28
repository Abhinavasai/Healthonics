import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import {
  BehaviorSubject,
  Observable,
  Subject,
  catchError,
  map,
  of,
  tap
} from 'rxjs';
import { ApiContract } from './api-contract';
import { AuthService } from './auth.service';

/** Aligns with Sprint 3 backend: GET/POST /api/messages/... */
export interface MessageThread {
  id: string;
  peer_user_id: string;
  peer_email: string;
  last_message_at: string;
  last_preview: string;
  unread_count: number;
}

export interface ChatMessage {
  id: string;
  thread_id: string;
  sender_id: string;
  body: string;
  created_at: string;
}

export interface ThreadsResponse {
  threads: MessageThread[];
}

export interface MessagesResponse {
  messages: ChatMessage[];
}

export interface UnreadResponse {
  unread_total: number;
}

/** Payload when backend pushes `new_message` over WebSocket (PR-22+). */
export interface ChatMessageRealtimeEvent {
  thread_id: string;
  message: ChatMessage;
}

export type MessagingRealtimeState = 'offline' | 'connecting' | 'live';

/** Builds ws(s) URL for GET /api/messages/ws — exposed for unit tests. */
export function messagingWebSocketUrl(token: string): string {
  const proto =
    typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost';
  return `${proto}//${host}/api/messages/ws?token=${encodeURIComponent(token)}`;
}

@Injectable({ providedIn: 'root' })
export class MessagingService {
  private readonly unreadSubject = new BehaviorSubject<number>(0);
  readonly unread$ = this.unreadSubject.asObservable();

  private ws: WebSocket | null = null;

  private readonly newMessageSubject = new Subject<ChatMessageRealtimeEvent>();
  readonly newChatMessage$ = this.newMessageSubject.asObservable();

  private readonly connectionStateSubject = new BehaviorSubject<MessagingRealtimeState>('offline');
  readonly connectionState$ = this.connectionStateSubject.asObservable();

  constructor(
    private http: HttpClient,
    private auth: AuthService
  ) {}

  getUnreadSnapshot(): number {
    return this.unreadSubject.value;
  }

  /** Preferred when backend exposes aggregate unread. */
  refreshUnread(): Observable<number> {
    return this.http.get<UnreadResponse>(ApiContract.messaging.unread).pipe(
      map((r) => r.unread_total ?? 0),
      tap((n) => this.unreadSubject.next(n)),
      catchError(() => {
        this.unreadSubject.next(0);
        return of(0);
      })
    );
  }

  /** Fallback: derive unread from thread rows when /unread is missing. */
  applyUnreadFromThreads(threads: MessageThread[]): void {
    const sum = (threads ?? []).reduce((acc, t) => acc + (t.unread_count ?? 0), 0);
    this.unreadSubject.next(sum);
  }

  /**
   * Opens a WebSocket to `/api/messages/ws?token=…` when a JWT exists.
   * Safe no-op without token, duplicate connect, or missing WebSocket API.
   * Falls back silently if the backend route is unavailable (polling still works).
   */
  connectRealtime(): void {
    if (typeof WebSocket === 'undefined') {
      return;
    }
    const token = this.auth.getToken();
    if (!token) {
      return;
    }
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return;
    }

    this.disconnectRealtime();
    this.connectionStateSubject.next('connecting');

    let socket: WebSocket;
    try {
      socket = new WebSocket(messagingWebSocketUrl(token));
    } catch {
      this.connectionStateSubject.next('offline');
      return;
    }

    this.ws = socket;

    socket.onopen = () => {
      this.connectionStateSubject.next('live');
    };

    socket.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(String(event.data)) as Record<string, unknown>;
        if (data['type'] !== 'new_message') {
          return;
        }
        const threadId = data['thread_id'];
        const rawMsg = data['message'];
        if (typeof threadId !== 'string' || typeof rawMsg !== 'object' || rawMsg === null) {
          return;
        }
        const m = rawMsg as Record<string, unknown>;
        const chat: ChatMessage = {
          id: String(m['id'] ?? ''),
          thread_id: String(m['thread_id'] ?? threadId),
          sender_id: String(m['sender_id'] ?? ''),
          body: String(m['body'] ?? ''),
          created_at: String(m['created_at'] ?? '')
        };
        if (!chat.id) {
          return;
        }
        this.newMessageSubject.next({ thread_id: threadId, message: chat });
      } catch {
        /* ignore malformed frames */
      }
    };

    socket.onerror = () => {
      if (this.connectionStateSubject.value === 'connecting') {
        this.connectionStateSubject.next('offline');
      }
    };

    socket.onclose = () => {
      if (this.ws === socket) {
        this.ws = null;
      }
      this.connectionStateSubject.next('offline');
    };
  }

  disconnectRealtime(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connectionStateSubject.next('offline');
  }

  listThreads(): Observable<MessageThread[]> {
    return this.http.get<ThreadsResponse>(ApiContract.messaging.threads).pipe(
      map((r) => r.threads ?? []),
      tap((threads) => this.applyUnreadFromThreads(threads)),
      catchError(() => {
        this.unreadSubject.next(0);
        return of([]);
      })
    );
  }

  listMessages(threadId: string): Observable<ChatMessage[]> {
    return this.http
      .get<MessagesResponse>(ApiContract.messaging.threadMessages(encodeURIComponent(threadId)))
      .pipe(
        map((r) => r.messages ?? []),
        catchError(() => of([]))
      );
  }

  sendMessage(threadId: string, body: string): Observable<ChatMessage> {
    return this.http.post<ChatMessage>(ApiContract.messaging.sendMessage(encodeURIComponent(threadId)), {
      body
    });
  }

  createThread(peerUserId: string, body: string): Observable<MessageThread> {
    return this.http.post<MessageThread>(ApiContract.messaging.threads, { peer_user_id: peerUserId, body });
  }

  markThreadRead(threadId: string): Observable<void> {
    return this.http
      .post<void>(ApiContract.messaging.markRead(encodeURIComponent(threadId)), {})
      .pipe(catchError(() => of(undefined)));
  }
}

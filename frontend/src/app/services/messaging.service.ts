import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, catchError, map, of, tap } from 'rxjs';
import { ApiContract } from './api-contract';

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

@Injectable({ providedIn: 'root' })
export class MessagingService {
  private readonly unreadSubject = new BehaviorSubject<number>(0);
  readonly unread$ = this.unreadSubject.asObservable();

  constructor(private http: HttpClient) {}

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

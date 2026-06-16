import { Component, ElementRef, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { finalize, Subscription, switchMap, timer } from 'rxjs';
import {
  ChatMessage,
  MessageThread,
  MessagingService
} from '../../services/messaging.service';
import { AuthService } from '../../services/auth.service';
import { AppointmentsService } from '../../services/appointments.service';
import { AssistantService } from '../../services/assistant.service';

/** Inbox + thread view for patients and doctors (Sprint 3 feature 9). */
@Component({
  selector: 'app-messages',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, DatePipe],
  templateUrl: './messages.component.html',
  styleUrl: './messages.component.scss'
})
export class MessagesComponent implements OnInit, OnDestroy {
  @ViewChild('chatBody') chatBody?: ElementRef<HTMLDivElement>;

  threads: MessageThread[] = [];
  messages: ChatMessage[] = [];
  selectedThreadId: string | null = null;
  loadingThreads = false;
  loadingMessages = false;
  sending = false;
  error = '';
  replyBody = '';
  newThreadBody = '';
  /** Patient: pick a doctor to start a thread. */
  doctors: { id: string; email: string }[] = [];
  selectedDoctorId = '';
  /** Doctor: patient user id for first message (MVP until "my patients" API exists). */
  peerPatientId = '';
  showNewThread = false;
  drafting = false;

  private routeSub?: Subscription;
  private pollSub?: Subscription;
  private rtMsgSub?: Subscription;
  private rtStateSub?: Subscription;
  /** When true, WebSocket is connected; polling becomes a slow safety net only. */
  realtimeLive = false;
  private readonly pollMs = 25000;
  private readonly pollFallbackMs = 120000;

  constructor(
    private messaging: MessagingService,
    private assistant: AssistantService,
    private auth: AuthService,
    private appointments: AppointmentsService,
    private route: ActivatedRoute,
    private router: Router
  ) {}

  get role(): string {
    return this.auth.getUser()?.role ?? 'patient';
  }

  get basePath(): string {
    return this.role === 'doctor' ? '/doctor/messages' : '/patient/messages';
  }

  get myUserId(): string {
    return this.auth.getUser()?.id ?? '';
  }

  get selectedPeerEmail(): string {
    const t = this.threads.find((x) => x.id === this.selectedThreadId);
    return t?.peer_email ?? 'Conversation';
  }

  ngOnInit(): void {
    if (this.role === 'patient') {
      this.appointments.listDoctors().subscribe({
        next: (res) => (this.doctors = res.doctors ?? []),
        error: () => (this.doctors = [])
      });
    }

    this.routeSub = this.route.paramMap.subscribe((params) => {
      const id = params.get('threadId');
      this.selectedThreadId = id;
      this.loadThreads();
      if (id) {
        this.loadMessages(id);
        this.startPollingThread(id);
      } else {
        this.messages = [];
        this.stopPolling();
      }
    });

    this.messaging.connectRealtime();

    this.rtStateSub = this.messaging.connectionState$.subscribe((s) => {
      this.realtimeLive = s === 'live';
      const tid = this.selectedThreadId;
      if (tid) {
        this.startPollingThread(tid);
      }
    });

    this.rtMsgSub = this.messaging.newChatMessage$.subscribe((evt) => {
      this.loadThreads();
      if (evt.thread_id !== this.selectedThreadId || !evt.message) {
        return;
      }
      const m = evt.message;
      if (this.messages.some((x) => x.id === m.id)) {
        return;
      }
      this.messages = [...this.messages, m].sort(
        (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      );
      this.scrollChatToEnd();
      this.messaging.markThreadRead(evt.thread_id).subscribe(() => this.loadThreads());
    });
  }

  ngOnDestroy(): void {
    this.routeSub?.unsubscribe();
    this.rtMsgSub?.unsubscribe();
    this.rtStateSub?.unsubscribe();
    this.stopPolling();
    this.messaging.disconnectRealtime();
  }

  loadThreads(): void {
    this.loadingThreads = true;
    this.error = '';
    this.messaging
      .listThreads()
      .pipe(finalize(() => (this.loadingThreads = false)))
      .subscribe({
        next: (list) => (this.threads = list),
        error: (err) => {
          this.threads = [];
          this.error = err?.error?.error ?? 'Unable to load conversations.';
        }
      });
  }

  loadMessages(threadId: string): void {
    this.loadingMessages = true;
    this.messaging
      .listMessages(threadId)
      .pipe(finalize(() => (this.loadingMessages = false)))
      .subscribe({
        next: (list) => {
          this.messages = [...list].sort(
            (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
          );
          this.scrollChatToEnd();
          this.messaging.markThreadRead(threadId).subscribe(() => this.loadThreads());
        },
        error: () => {
          this.messages = [];
        }
      });
  }

  private startPollingThread(threadId: string): void {
    this.stopPolling();
    const interval = this.realtimeLive ? this.pollFallbackMs : this.pollMs;
    this.pollSub = timer(interval, interval)
      .pipe(switchMap(() => this.messaging.listMessages(threadId)))
      .subscribe({
        next: (list) => {
          this.messages = [...list].sort(
            (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
          );
          this.scrollChatToEnd();
        },
        error: (err) => {
          console.error('Message polling error, will retry on next interval:', err);
          // Restart polling after a delay so transient errors don't kill the feed.
          this.stopPolling();
          setTimeout(() => {
            if (this.selectedThreadId === threadId) {
              this.startPollingThread(threadId);
            }
          }, 10000);
        }
      });
  }

  private stopPolling(): void {
    this.pollSub?.unsubscribe();
    this.pollSub = undefined;
  }

  selectThread(thread: MessageThread): void {
    this.router.navigate([this.basePath, thread.id]);
  }

  send(): void {
    const tid = this.selectedThreadId;
    const text = this.replyBody.trim();
    if (!tid || !text || this.sending) {
      return;
    }
    this.sending = true;
    this.messaging
      .sendMessage(tid, text)
      .pipe(finalize(() => (this.sending = false)))
      .subscribe({
        next: () => {
          this.replyBody = '';
          this.loadMessages(tid);
          this.loadThreads();
          this.messaging.refreshUnread().subscribe();
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to send message.';
        }
      });
  }

  generateReplyDraft(): void {
    if (!this.selectedThreadId || this.drafting) {
      return;
    }
    const recent = this.messages.slice(-6);
    const context = recent
      .map((m) => `${this.isMine(m) ? 'me' : 'peer'}: ${m.body}`)
      .join('\n');
    this.drafting = true;
    // Pass conversation context as user/assistant history turns.
    const history = recent.map((m) => ({
      role: (this.isMine(m) ? 'user' : 'assistant') as 'user' | 'assistant',
      content: m.body
    }));
    this.assistant.chat('Draft a short professional reply for this conversation.', history).subscribe({
      next: (res) => {
        this.drafting = false;
        this.replyBody = (res.reply || '').trim();
      },
      error: () => {
        this.drafting = false;
        this.error = 'Could not generate AI draft right now.';
      }
    });
  }

  createThread(): void {
    const body = this.newThreadBody.trim();
    const peer =
      this.role === 'patient' ? this.selectedDoctorId.trim() : this.peerPatientId.trim();
    if (!peer || !body) {
      this.error = 'Choose a recipient and enter a message.';
      return;
    }
    this.sending = true;
    this.error = '';
    this.messaging
      .createThread(peer, body)
      .pipe(finalize(() => (this.sending = false)))
      .subscribe({
        next: (t) => {
          this.newThreadBody = '';
          this.peerPatientId = '';
          this.showNewThread = false;
          this.router.navigate([this.basePath, t.id]);
          this.messaging.refreshUnread().subscribe();
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to start conversation.';
        }
      });
  }

  isMine(msg: ChatMessage): boolean {
    return msg.sender_id === this.myUserId;
  }

  private scrollChatToEnd(): void {
    queueMicrotask(() => {
      const el = this.chatBody?.nativeElement;
      if (el) {
        el.scrollTop = el.scrollHeight;
      }
    });
  }
}

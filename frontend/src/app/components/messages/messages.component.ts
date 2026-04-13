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

/** Inbox + thread view for patients and doctors (Sprint 3 feature 9 + F10 limits / compose UX). */
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

  private routeSub?: Subscription;
  private pollSub?: Subscription;
  private readonly pollMs = 25000;

  /** Matches backend `maxMessageRunes` in messaging handler. */
  maxMessageRunes = 8000;

  constructor(
    private messaging: MessagingService,
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

  /** Unicode-aware length for parity with server-side rune cap. */
  runeLen(s: string): number {
    return [...(s || '')].length;
  }

  get newThreadRunes(): number {
    return this.runeLen(this.newThreadBody);
  }

  get replyRunes(): number {
    return this.runeLen(this.replyBody);
  }

  get newThreadOverLimit(): boolean {
    return this.newThreadRunes > this.maxMessageRunes;
  }

  get replyOverLimit(): boolean {
    return this.replyRunes > this.maxMessageRunes;
  }

  ngOnInit(): void {
    this.messaging.composeLimits().subscribe({
      next: (limits) => {
        if (limits?.max_body_runes && limits.max_body_runes > 0) {
          this.maxMessageRunes = limits.max_body_runes;
        }
      }
    });
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
  }

  ngOnDestroy(): void {
    this.routeSub?.unsubscribe();
    this.stopPolling();
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
    this.pollSub = timer(this.pollMs, this.pollMs)
      .pipe(switchMap(() => this.messaging.listMessages(threadId)))
      .subscribe({
        next: (list) => {
          this.messages = [...list].sort(
            (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
          );
          this.scrollChatToEnd();
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
    if (this.replyOverLimit) {
      this.error = `Message is too long (max ${this.maxMessageRunes} characters).`;
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

  createThread(): void {
    const body = this.newThreadBody.trim();
    const peer =
      this.role === 'patient' ? this.selectedDoctorId.trim() : this.peerPatientId.trim();
    if (!peer || !body) {
      this.error = 'Choose a recipient and enter a message.';
      return;
    }
    if (this.newThreadOverLimit) {
      this.error = `First message is too long (max ${this.maxMessageRunes} characters).`;
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

import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, TitleCasePipe, AsyncPipe } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { Subscription, timer, switchMap } from 'rxjs';
import { AuthService } from '../../services/auth.service';
import { MessagingService } from '../../services/messaging.service';
import { NotificationsInboxService, NotificationRow } from '../../services/notifications.service';
import { FormsModule } from '@angular/forms';
import { AssistantService, ChatMessage } from '../../services/assistant.service';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe, AsyncPipe, FormsModule],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss'
})
export class AppShellComponent implements OnInit, OnDestroy {
  private unreadPoll?: Subscription;
  private notifPoll?: Subscription;
  private realtimeUnreadSub?: Subscription;
  private seenNotificationIds = new Set<string>();
  notificationPopup: NotificationRow | null = null;
  assistantOpen = false;
  assistantInput = '';
  assistantSending = false;
  assistantMessages: { from: 'assistant' | 'me'; text: string; meta?: string }[] = [
    { from: 'assistant', text: 'Hi! I\'m your Healthonyx Care Assistant — powered by AI. I can help with health questions, appointments, prescriptions, lab results, documents, finding care, and anything else on your mind. What can I help you with?' }
  ];
  /** Canonical history for the backend — only user/assistant turns, no system noise. */
  private conversationHistory: ChatMessage[] = [];

  constructor(
    public auth: AuthService,
    public messaging: MessagingService,
    private notifications: NotificationsInboxService,
    private assistant: AssistantService,
    private router: Router
  ) {}

  ngOnInit(): void {
    const role = this.user?.role;
    if (role === 'patient' || role === 'doctor') {
      this.messaging.connectRealtime();
      this.realtimeUnreadSub = this.messaging.newChatMessage$.subscribe(() => {
        this.messaging.refreshUnread().subscribe();
      });
    }

    // Only poll unread messages for roles that have messaging access
    if (role === 'patient' || role === 'doctor') {
      this.unreadPoll = timer(0, 30_000)
        .pipe(switchMap(() => this.messaging.refreshUnread()))
        .subscribe();
    }
    if (role === 'patient' || role === 'doctor') {
      this.notifPoll = timer(0, 20_000).pipe(switchMap(() => this.notifications.listMine())).subscribe({
        next: (res) => this.handleNotificationPoll(res.notifications ?? [])
      });
    }
  }

  ngOnDestroy(): void {
    this.unreadPoll?.unsubscribe();
    this.notifPoll?.unsubscribe();
    this.realtimeUnreadSub?.unsubscribe();
    this.messaging.disconnectRealtime();
  }

  get user() {
    return this.auth.getUser();
  }

  get role() {
    return this.user?.role ?? 'patient';
  }

  logout(): void {
    this.auth.logout();
  }

  /** False for routes with detail children (appointments/:id, documents/:documentId, messages/:threadId). */
  isExactNavPath(path: string): boolean {
    if (
      path.endsWith('/appointments') ||
      path.endsWith('/messages') ||
      path.endsWith('/prescriptions') ||
      path.endsWith('/notifications') ||
      path.endsWith('/dashboard') ||
      path.endsWith('/knowledge') ||
      path.endsWith('/audit') ||
      path.endsWith('/documents')
    ) {
      return false;
    }
    return true;
  }

  get navLinks(): { path: string; label: string; roles: string[] }[] {
    const all = [
      { path: '/patient/dashboard', label: 'Dashboard', roles: ['patient'] },
      { path: '/patient/find-care', label: 'Find care', roles: ['patient'] },
      { path: '/patient/prescriptions', label: 'Prescriptions', roles: ['patient'] },
      { path: '/patient/appointments', label: 'My Appointments', roles: ['patient'] },
      { path: '/patient/documents', label: 'My documents', roles: ['patient'] },
      { path: '/patient/files', label: 'My files', roles: ['patient'] },
      { path: '/patient/messages', label: 'Messages', roles: ['patient'] },
      { path: '/patient/notifications', label: 'Notifications', roles: ['patient'] },
      { path: '/patient/symptom-check', label: 'Symptom Checker', roles: ['patient'] },
      { path: '/doctor/dashboard', label: 'Dashboard', roles: ['doctor'] },
      { path: '/doctor/availability', label: 'Availability', roles: ['doctor'] },
      { path: '/doctor/appointments', label: 'Appointment Queue', roles: ['doctor'] },
      { path: '/doctor/documents', label: 'Patient documents', roles: ['doctor'] },
      { path: '/doctor/messages', label: 'Messages', roles: ['doctor'] },
      { path: '/doctor/notifications', label: 'Notifications', roles: ['doctor'] },
      { path: '/doctor/second-opinions', label: 'Second Opinions', roles: ['doctor'] },
      { path: '/doctor/patients', label: 'My Patients', roles: ['doctor'] },
      { path: '/doctor/waiting-room', label: 'Waiting Room', roles: ['doctor'] },
      { path: '/admin', label: 'Admin Dashboard', roles: ['admin'] },
      { path: '/admin/audit', label: 'Audit log', roles: ['admin'] },
      { path: '/admin/knowledge', label: 'Knowledge (admin)', roles: ['admin'] },
      { path: '/admin/slo', label: 'SLO Dashboard', roles: ['admin'] }
    ];
    return all.filter((l) => l.roles.includes(this.role));
  }

  dismissNotificationPopup(): void {
    this.notificationPopup = null;
  }

  toggleAssistant(): void {
    this.assistantOpen = !this.assistantOpen;
  }

  sendAssistant(): void {
    const q = this.assistantInput.trim();
    if (!q || this.assistantSending) {
      return;
    }
    this.assistantMessages = [...this.assistantMessages, { from: 'me', text: q }];
    this.assistantInput = '';
    this.assistantSending = true;

    // Send full conversation history for multi-turn context.
    this.assistant.chat(q, this.conversationHistory).subscribe({
      next: (res) => {
        this.assistantSending = false;
        const reply = res.reply?.trim() || 'Sorry, I could not generate a response.';
        this.assistantMessages = [
          ...this.assistantMessages,
          { from: 'assistant', text: reply, meta: res.fallback_used ? undefined : `via ${res.provider}` }
        ];
        // Append both turns to history for next request.
        this.conversationHistory = [
          ...this.conversationHistory,
          { role: 'user', content: q },
          { role: 'assistant', content: reply }
        ];
      },
      error: () => {
        this.assistantSending = false;
        this.assistantMessages = [
          ...this.assistantMessages,
          { from: 'assistant', text: 'Assistant is temporarily unavailable. Please try again.' }
        ];
      }
    });
  }

  clearConversation(): void {
    this.conversationHistory = [];
    this.assistantMessages = [
      { from: 'assistant', text: 'Conversation cleared. How can I help you?' }
    ];
  }

  private handleNotificationPoll(rows: NotificationRow[]): void {
    if (rows.length === 0) {
      return;
    }
    if (this.seenNotificationIds.size === 0) {
      rows.forEach((r) => this.seenNotificationIds.add(r.id));
      return;
    }
    const newest = rows.find((r) => !this.seenNotificationIds.has(r.id));
    rows.forEach((r) => this.seenNotificationIds.add(r.id));
    if (newest) {
      this.notificationPopup = newest;
    }
  }

}

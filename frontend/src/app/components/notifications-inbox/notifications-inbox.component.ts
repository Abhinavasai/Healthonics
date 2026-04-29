import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  NotificationsInboxService,
  NotificationPreferenceRow,
  NotificationRow
} from '../../services/notifications.service';
import { Subscription, timer } from 'rxjs';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-notifications-inbox',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section data-cy="notifications-inbox-page">
      <h1>Notifications</h1>
      <p class="ok" *ngIf="prefsSaved">{{ prefsSaved }}</p>
      <p class="err" *ngIf="error">{{ error }}</p>

      <section class="prefs" data-cy="notification-preferences">
        <h2>Preferences</h2>
        <p class="muted">Choose delivery channels per category and per item.</p>

        <div class="prefs-toolbar" *ngIf="preferences.length">
          <span class="muted">Bulk channel controls:</span>
          <button type="button" class="ghost" (click)="setAll('email_enabled', true)" data-cy="pref-bulk-email-on">
            Email all
          </button>
          <button type="button" class="ghost" (click)="setAll('sms_enabled', true)" data-cy="pref-bulk-sms-on">
            SMS all
          </button>
          <button type="button" class="ghost" (click)="setAll('in_app_enabled', true)" data-cy="pref-bulk-inapp-on">
            In-app all
          </button>
          <button type="button" class="ghost" (click)="setAll('enabled', false)" data-cy="pref-bulk-disable-all">
            Disable all
          </button>
        </div>

        <div class="prefs-cards" *ngIf="preferences.length; else prefsFallback">
          <div class="pref-card" *ngFor="let p of preferences; let i = index" data-cy="notification-pref-row">
            <div class="pref-head">
              <strong>{{ labelFor(p.category) }}</strong>
              <label class="toggle">
                <input
                  type="checkbox"
                  [checked]="p.enabled"
                  (change)="setCategoryEnabled(i, $event)"
                  data-cy="pref-category-enabled"
                />
                <span>{{ p.enabled ? 'Enabled' : 'Disabled' }}</span>
              </label>
            </div>
            <div class="pref-items">
              <label class="toggle">
                <input
                  type="checkbox"
                  [checked]="p.email_enabled"
                  [disabled]="!p.enabled"
                  (change)="setChannel(i, 'email_enabled', $event)"
                  data-cy="pref-item-email"
                />
                <span>Email</span>
              </label>
              <label class="toggle">
                <input
                  type="checkbox"
                  [checked]="p.sms_enabled"
                  [disabled]="!p.enabled"
                  (change)="setChannel(i, 'sms_enabled', $event)"
                  data-cy="pref-item-sms"
                />
                <span>SMS</span>
              </label>
              <label class="toggle">
                <input
                  type="checkbox"
                  [checked]="p.in_app_enabled"
                  [disabled]="!p.enabled"
                  (change)="setChannel(i, 'in_app_enabled', $event)"
                  data-cy="pref-item-inapp"
                />
                <span>In-app</span>
              </label>
            </div>
          </div>
        </div>
        <ng-template #prefsFallback>
          <p class="muted">Preferences are unavailable right now.</p>
        </ng-template>
        <button type="button" (click)="savePreferences()" [disabled]="saving || !preferences.length" data-cy="notification-pref-save">
          {{ saving ? 'Saving...' : 'Save preferences' }}
        </button>
      </section>

      <div class="filters" *ngIf="rows.length">
        <button type="button" class="ghost" (click)="setFilter('all')">All</button>
        <button type="button" class="ghost" (click)="setFilter('pending')">Pending</button>
        <button type="button" class="ghost" (click)="setFilter('sent')">Sent</button>
        <button type="button" class="ghost" (click)="setFilter('failed')">Failed</button>
      </div>

      <ul *ngIf="filteredRows.length" class="notif-list" data-cy="notifications-list">
        <li *ngFor="let r of filteredRows" class="notif-row" data-cy="notifications-row">
          <div class="notif-row__head">
            <strong>{{ r.title }}</strong>
            <span class="pill">{{ r.status | titlecase }}</span>
          </div>
          <div class="body">{{ r.body }}</div>
          <small>{{ r.channel | titlecase }} · {{ formatTimestamp(r.created_at) }}</small>
          <div class="notif-actions">
            <button type="button" class="ghost" (click)="openNotification(r)">Open</button>
          </div>
        </li>
      </ul>
      <p *ngIf="!error && !filteredRows.length" data-cy="notifications-empty">No notifications.</p>
    </section>
  `,
  styles: [
    `
      .err {
        color: #b00020;
      }
      .ok {
        color: #166534;
      }
      .body {
        margin: 0.25rem 0;
      }
      .prefs {
        margin: 1rem 0 1.25rem;
        padding: 0.75rem;
        border: 1px solid #e5e7eb;
        border-radius: 0.5rem;
      }
      .muted {
        color: #6b7280;
      }
      .prefs-toolbar {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: 0.6rem;
      }
      .prefs-cards {
        display: grid;
        gap: 0.5rem;
      }
      .pref-card {
        border: 1px solid #e5e7eb;
        border-radius: 0.5rem;
        padding: 0.6rem;
        display: grid;
        gap: 0.5rem;
      }
      .pref-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.5rem;
      }
      .pref-items {
        display: flex;
        flex-wrap: wrap;
        gap: 0.75rem;
      }
      .toggle {
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
      }
      .filters {
        display: flex;
        gap: 0.45rem;
        flex-wrap: wrap;
        margin: 0.45rem 0 0.75rem;
      }
      .notif-list {
        list-style: none;
        margin: 0;
        padding: 0;
        display: grid;
        gap: 0.6rem;
      }
      .notif-row {
        border: 1px solid #334155;
        border-radius: 12px;
        padding: 0.65rem 0.75rem;
        background: rgba(15, 23, 42, 0.65);
      }
      .notif-row__head {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 0.5rem;
      }
      .pill {
        border: 1px solid #0ea5e9;
        color: #22d3ee;
        border-radius: 999px;
        font-size: 0.72rem;
        padding: 0.1rem 0.5rem;
      }
      .notif-actions {
        margin-top: 0.45rem;
        display: flex;
        justify-content: flex-end;
      }
      .ghost {
        margin-top: 0;
        border: 1px solid #d1d5db;
        background: #fff;
        border-radius: 999px;
        padding: 0.2rem 0.55rem;
        cursor: pointer;
      }
      button {
        margin-top: 0.75rem;
      }
    `
  ]
})
export class NotificationsInboxComponent implements OnInit, OnDestroy {
  readonly categoryOrder = [
    'appointment_reminders',
    'medication_reminders',
    'lab_result_alerts',
    'announcements'
  ];

  rows: NotificationRow[] = [];
  filteredRows: NotificationRow[] = [];
  preferences: NotificationPreferenceRow[] = [];
  error: string | null = null;
  prefsSaved: string | null = null;
  saving = false;
  activeFilter: 'all' | 'pending' | 'failed' | 'sent' = 'all';
  private pollSub?: Subscription;

  constructor(
    private api: NotificationsInboxService,
    private router: Router,
    private auth: AuthService
  ) {}

  ngOnInit(): void {
    this.loadPreferences();
    this.loadNotifications();
    this.pollSub = timer(15000, 15000).subscribe(() => this.loadNotifications());
  }

  ngOnDestroy(): void {
    this.pollSub?.unsubscribe();
  }

  private loadPreferences(): void {
    this.api.listPreferences().subscribe({
      next: (res) => {
        this.preferences = this.sortPreferences(res.preferences ?? []);
      },
      error: () => {
        this.preferences = this.sortPreferences([
          { category: 'appointment_reminders', enabled: true, email_enabled: true, sms_enabled: false, in_app_enabled: true },
          { category: 'medication_reminders', enabled: true, email_enabled: true, sms_enabled: false, in_app_enabled: true },
          { category: 'lab_result_alerts', enabled: true, email_enabled: true, sms_enabled: false, in_app_enabled: true },
          { category: 'announcements', enabled: true, email_enabled: true, sms_enabled: false, in_app_enabled: true }
        ]);
      }
    });
  }

  private sortPreferences(rows: NotificationPreferenceRow[]): NotificationPreferenceRow[] {
    const rank = new Map(this.categoryOrder.map((v, i) => [v, i]));
    return [...rows].sort((a, b) => {
      const aRank = rank.get(a.category) ?? Number.MAX_SAFE_INTEGER;
      const bRank = rank.get(b.category) ?? Number.MAX_SAFE_INTEGER;
      if (aRank !== bRank) {
        return aRank - bRank;
      }
      return a.category.localeCompare(b.category);
    });
  }

  labelFor(category: string): string {
    switch (category) {
      case 'appointment_reminders':
        return 'Appointment reminders';
      case 'medication_reminders':
        return 'Medication reminders';
      case 'lab_result_alerts':
        return 'Lab result alerts';
      case 'announcements':
        return 'Announcements';
      default:
        return category;
    }
  }

  setCategoryEnabled(index: number, event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input || !this.preferences[index]) {
      return;
    }
    const nextEnabled = input.checked;
    const current = this.preferences[index];
    this.preferences[index] = nextEnabled
      ? { ...current, enabled: true }
      : {
          ...current,
          enabled: false,
          email_enabled: false,
          sms_enabled: false,
          in_app_enabled: false
        };
    this.prefsSaved = null;
  }

  setChannel(index: number, key: 'email_enabled' | 'sms_enabled' | 'in_app_enabled', event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input || !this.preferences[index]) {
      return;
    }
    const next = { ...this.preferences[index], [key]: input.checked };
    next.enabled = Boolean(next.email_enabled || next.sms_enabled || next.in_app_enabled);
    this.preferences[index] = next;
    this.prefsSaved = null;
  }

  setAll(
    key: 'enabled' | 'email_enabled' | 'sms_enabled' | 'in_app_enabled',
    value: boolean
  ): void {
    this.preferences = this.preferences.map((p) => {
      if (key === 'enabled') {
        if (value) {
          return {
            ...p,
            enabled: true,
            email_enabled: true,
            sms_enabled: true,
            in_app_enabled: true
          };
        }
        return {
          ...p,
          enabled: false,
          email_enabled: false,
          sms_enabled: false,
          in_app_enabled: false
        };
      }
      const next = { ...p, enabled: true, [key]: value };
      next.enabled = Boolean(next.email_enabled || next.sms_enabled || next.in_app_enabled);
      return next;
    });
    this.prefsSaved = null;
  }

  savePreferences(): void {
    if (!this.preferences.length || this.saving) {
      return;
    }
    this.error = null;
    this.saving = true;
    this.api.savePreferences(this.preferences).subscribe({
      next: () => {
        this.saving = false;
        this.prefsSaved = 'Preferences saved.';
      },
      error: () => {
        this.saving = false;
        this.error = 'Could not save notification preferences';
      }
    });
  }

  setFilter(next: 'all' | 'pending' | 'failed' | 'sent'): void {
    this.activeFilter = next;
    this.applyFilter();
  }

  formatTimestamp(raw: string): string {
    const t = new Date(raw);
    if (Number.isNaN(t.getTime())) {
      return raw;
    }
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(t);
  }

  private loadNotifications(): void {
    this.api.listMine().subscribe({
      next: (res) => {
        this.rows = res.notifications ?? [];
        this.applyFilter();
      },
      error: () => (this.error = 'Could not load notifications')
    });
  }

  private applyFilter(): void {
    this.filteredRows =
      this.activeFilter === 'all'
        ? [...this.rows]
        : this.rows.filter((r) => (r.status || '').toLowerCase() === this.activeFilter);
  }

  openNotification(r: NotificationRow): void {
    const role = this.auth.getUser()?.role === 'doctor' ? 'doctor' : 'patient';
    const text = `${r.title} ${r.body}`.toLowerCase();
    if (text.includes('appointment')) {
      void this.router.navigate([`/${role}/appointments`]);
      return;
    }
    if (text.includes('lab') || text.includes('document')) {
      void this.router.navigate([`/${role}/documents`]);
      return;
    }
    if (text.includes('medication') || text.includes('prescription')) {
      void this.router.navigate(['/patient/prescriptions']);
      return;
    }
    if (text.includes('message')) {
      void this.router.navigate([`/${role}/messages`]);
      return;
    }
    void this.router.navigate([`/${role}/dashboard`]);
  }
}

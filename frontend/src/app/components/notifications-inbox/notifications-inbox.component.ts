import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  NotificationsInboxService,
  NotificationPreferenceRow,
  NotificationRow
} from '../../services/notifications.service';

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
        <p class="muted">Choose how you receive reminders and alerts.</p>
        <div class="prefs-table" *ngIf="preferences.length; else prefsFallback">
          <div class="prefs-head">
            <span>Category</span>
            <span>Enabled</span>
            <span>Email</span>
            <span>SMS</span>
            <span>In-app</span>
          </div>
          <div class="prefs-row" *ngFor="let p of preferences; let i = index" data-cy="notification-pref-row">
            <span>{{ labelFor(p.category) }}</span>
            <input type="checkbox" [checked]="p.enabled" (change)="toggle(i, 'enabled', $event)" />
            <input type="checkbox" [checked]="p.email_enabled" (change)="toggle(i, 'email_enabled', $event)" />
            <input type="checkbox" [checked]="p.sms_enabled" (change)="toggle(i, 'sms_enabled', $event)" />
            <input type="checkbox" [checked]="p.in_app_enabled" (change)="toggle(i, 'in_app_enabled', $event)" />
          </div>
        </div>
        <ng-template #prefsFallback>
          <p class="muted">Preferences are unavailable right now.</p>
        </ng-template>
        <button type="button" (click)="savePreferences()" [disabled]="saving || !preferences.length" data-cy="notification-pref-save">
          {{ saving ? 'Saving...' : 'Save preferences' }}
        </button>
      </section>

      <ul *ngIf="rows.length" data-cy="notifications-list">
        <li *ngFor="let r of rows" data-cy="notifications-row">
          <strong>{{ r.title }}</strong>
          <div class="body">{{ r.body }}</div>
          <small>{{ r.status }} · {{ r.created_at }}</small>
        </li>
      </ul>
      <p *ngIf="!error && !rows.length" data-cy="notifications-empty">No notifications.</p>
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
      .prefs-table {
        display: grid;
        gap: 0.5rem;
      }
      .prefs-head,
      .prefs-row {
        display: grid;
        grid-template-columns: 2fr repeat(4, minmax(64px, 0.5fr));
        align-items: center;
        gap: 0.5rem;
      }
      .prefs-head {
        font-weight: 600;
      }
      button {
        margin-top: 0.75rem;
      }
    `
  ]
})
export class NotificationsInboxComponent implements OnInit {
  rows: NotificationRow[] = [];
  preferences: NotificationPreferenceRow[] = [];
  error: string | null = null;
  prefsSaved: string | null = null;
  saving = false;

  constructor(private api: NotificationsInboxService) {}

  ngOnInit(): void {
    this.loadPreferences();
    this.api.listMine().subscribe({
      next: (res) => (this.rows = res.notifications ?? []),
      error: () => (this.error = 'Could not load notifications')
    });
  }

  private loadPreferences(): void {
    this.api.listPreferences().subscribe({
      next: (res) => (this.preferences = res.preferences ?? []),
      error: () => {
        this.preferences = [];
      }
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

  toggle(
    index: number,
    key: 'enabled' | 'email_enabled' | 'sms_enabled' | 'in_app_enabled',
    event: Event
  ): void {
    const input = event.target as HTMLInputElement | null;
    if (!input || !this.preferences[index]) {
      return;
    }
    this.preferences[index] = { ...this.preferences[index], [key]: input.checked };
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
}

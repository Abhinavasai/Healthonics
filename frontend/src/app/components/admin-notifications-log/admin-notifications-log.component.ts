import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  AdminNotificationRow,
  AdminNotificationsSummary,
  NotificationConsentEventRow,
  NotificationsInboxService
} from '../../services/notifications.service';

@Component({
  selector: 'app-admin-notifications-log',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './admin-notifications-log.component.html',
  styleUrl: './admin-notifications-log.component.scss'
})
export class AdminNotificationsLogComponent implements OnInit {
  rows: AdminNotificationRow[] = [];
  consentRows: NotificationConsentEventRow[] = [];
  summary: AdminNotificationsSummary | null = null;
  loading = true;
  retryingId: string | null = null;
  error: string | null = null;
  consentWarning: string | null = null;

  constructor(private api: NotificationsInboxService) {}

  ngOnInit(): void {
    this.load();
  }

  get providerCards(): Array<{ name: string; sent: number; failed: number; pending: number; retries: number; health: string }> {
    const providers = ['in_app', 'sendgrid', 'twilio'];
    return providers.map((providerName) => {
      const sourceRows =
        providerName === 'in_app'
          ? this.rows.filter((r) => r.channel === 'in_app')
          : this.rows.filter((r) => (r.provider ?? '').toLowerCase() === providerName);
      const sent = sourceRows.filter((r) => r.status === 'sent').length;
      const failed = sourceRows.filter((r) => r.status === 'failed').length;
      const pending = sourceRows.filter((r) => r.status === 'pending').length;
      const retries = sourceRows.filter((r) => (r.attempts ?? 0) > 1).length;
      const health = failed === 0 ? 'healthy' : failed > sent ? 'degraded' : 'warning';
      return { name: providerName.toUpperCase(), sent, failed, pending, retries, health };
    });
  }

  retry(row: AdminNotificationRow): void {
    this.retryingId = row.id;
    this.api.retryFailedNotification(row.id).subscribe({
      next: () => {
        this.retryingId = null;
        this.load();
      },
      error: () => {
        this.retryingId = null;
        this.error = 'Could not retry failed notification.';
      }
    });
  }

  private load(): void {
    this.loading = true;
    this.error = null;
    this.consentWarning = null;
    this.api.adminSummary().subscribe({
      next: (summary) => {
        this.summary = summary;
        this.api.listAdminNotifications().subscribe({
          next: (r) => {
            this.rows = r.notifications ?? [];
            this.api.listAdminConsentHistory().subscribe({
              next: (consent) => {
                this.consentRows = consent.events ?? [];
                this.loading = false;
              },
              error: () => {
                this.consentWarning = 'Consent history is currently unavailable.';
                this.loading = false;
              }
            });
          },
          error: () => {
            this.error = 'Could not load notification log.';
            this.loading = false;
          }
        });
      },
      error: () => {
        this.error = 'Could not load notification summary.';
        this.loading = false;
      }
    });
  }
}


import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  AdminNotificationRow,
  AdminNotificationsSummary,
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
  summary: AdminNotificationsSummary | null = null;
  loading = true;
  retryingId: string | null = null;
  error: string | null = null;

  constructor(private api: NotificationsInboxService) {}

  ngOnInit(): void {
    this.load();
  }

  get providerCards(): Array<{ name: string; sent: number; failed: number; pending: number; health: string }> {
    const channels = ['in_app', 'email', 'sms'];
    return channels.map((ch) => {
      const sent = this.rows.filter((r) => r.channel === ch && r.status === 'sent').length;
      const failed = this.rows.filter((r) => r.channel === ch && r.status === 'failed').length;
      const pending = this.rows.filter((r) => r.channel === ch && r.status === 'pending').length;
      const health = failed === 0 ? 'healthy' : failed > sent ? 'degraded' : 'warning';
      return { name: ch.toUpperCase(), sent, failed, pending, health };
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
    this.api.adminSummary().subscribe({
      next: (summary) => {
        this.summary = summary;
        this.api.listAdminNotifications().subscribe({
          next: (r) => {
            this.rows = r.notifications ?? [];
            this.loading = false;
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


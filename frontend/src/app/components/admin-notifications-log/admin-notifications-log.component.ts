import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationsInboxService, NotificationRow } from '../../services/notifications.service';

@Component({
  selector: 'app-admin-notifications-log',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section data-cy="admin-notifications-page">
      <h1>Notification log</h1>
      <p *ngIf="error" class="err">{{ error }}</p>
      <table *ngIf="rows.length" data-cy="admin-notifications-table">
        <thead>
          <tr>
            <th>When</th>
            <th>Title</th>
            <th>Channel</th>
            <th>Status</th>
            <th>Scheduled</th>
            <th>Body</th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let r of rows" data-cy="admin-notifications-row">
            <td>{{ r.created_at }}</td>
            <td>{{ r.title }}</td>
            <td>{{ r.channel }}</td>
            <td>{{ r.status }}</td>
            <td>{{ r.scheduled_for || '—' }}</td>
            <td>{{ r.body }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!error && !rows.length" data-cy="admin-notifications-empty">
        No notifications yet.
      </p>
    </section>
  `,
  styles: [
    `
      table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.85rem;
      }
      th,
      td {
        border-bottom: 1px solid rgba(148, 163, 184, 0.25);
        padding: 0.4rem;
        text-align: left;
      }
      .err {
        color: #f87171;
      }
    `
  ]
})
export class AdminNotificationsLogComponent implements OnInit {
  rows: NotificationRow[] = [];
  error: string | null = null;

  constructor(private api: NotificationsInboxService) {}

  ngOnInit(): void {
    this.api.listMine().subscribe({
      next: (r) => (this.rows = r.notifications ?? []),
      error: () => (this.error = 'Could not load notification log')
    });
  }
}


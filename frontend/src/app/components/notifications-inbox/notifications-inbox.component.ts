import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationsInboxService, NotificationRow } from '../../services/notifications.service';

@Component({
  selector: 'app-notifications-inbox',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section data-cy="notifications-inbox-page">
      <h1>Notifications</h1>
      <p class="err" *ngIf="error">{{ error }}</p>
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
      .body {
        margin: 0.25rem 0;
      }
    `
  ]
})
export class NotificationsInboxComponent implements OnInit {
  rows: NotificationRow[] = [];
  error: string | null = null;

  constructor(private api: NotificationsInboxService) {}

  ngOnInit(): void {
    this.api.listMine().subscribe({
      next: (res) => (this.rows = res.notifications ?? []),
      error: () => (this.error = 'Could not load notifications')
    });
  }
}

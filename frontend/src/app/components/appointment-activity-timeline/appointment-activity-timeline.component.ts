import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { AppointmentActivity } from '../../services/appointments.service';

@Component({
  selector: 'app-appointment-activity-timeline',
  standalone: true,
  imports: [CommonModule, DatePipe],
  templateUrl: './appointment-activity-timeline.component.html',
  styleUrl: './appointment-activity-timeline.component.scss'
})
export class AppointmentActivityTimelineComponent {
  @Input() loading = false;
  @Input() refreshing = false;
  @Input() error = '';
  @Input() activities: AppointmentActivity[] = [];

  @Output() refreshRequested = new EventEmitter<void>();

  expandedId: string | null = null;

  headline(a: AppointmentActivity): string {
    switch (a.action) {
      case 'status_changed':
        if (a.detail === 'approved') {
          return 'Request approved';
        }
        if (a.detail === 'rejected') {
          return 'Request rejected';
        }
        return a.detail ? `Status updated (${a.detail})` : 'Status updated';
      default:
        return a.action.replace(/_/g, ' ');
    }
  }

  subline(a: AppointmentActivity): string {
    const em = a.actor_email?.trim();
    if (em) {
      return `By ${em}`;
    }
    const short = a.actor_user_id?.slice(0, 8) ?? '';
    return short ? `Actor ${short}…` : '';
  }

  onRefresh(): void {
    if (this.refreshing) {
      return;
    }
    this.refreshRequested.emit();
  }

  toggleExpand(id: string): void {
    this.expandedId = this.expandedId === id ? null : id;
  }

  isExpanded(a: AppointmentActivity): boolean {
    return this.expandedId === a.id;
  }

  hasExpandable(a: AppointmentActivity): boolean {
    return !!(a.detail || a.action);
  }
}

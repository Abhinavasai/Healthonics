import { Component, Input } from '@angular/core';
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
  @Input() error = '';
  @Input() activities: AppointmentActivity[] = [];

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
    if (a.actor_email) {
      return `Actor ${a.actor_email}`;
    }
    const short = a.actor_user_id?.slice(0, 8) ?? '';
    return short ? `Actor ${short}…` : '';
  }
}

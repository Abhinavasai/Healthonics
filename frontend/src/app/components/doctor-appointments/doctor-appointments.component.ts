import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { finalize } from 'rxjs';
import { Appointment, AppointmentsService } from '../../services/appointments.service';

@Component({
  selector: 'app-doctor-appointments',
  standalone: true,
  imports: [CommonModule, DatePipe, TitleCasePipe],
  template: `
    <section class="appointments">
      <h1>Doctor Appointments</h1>
      <p class="subtitle">Review incoming requests and record your decision.</p>

      <div class="card list">
        <div class="list-header">
          <h2>Assigned Requests</h2>
          <button type="button" (click)="load()" [disabled]="loading">
            {{ loading ? 'Refreshing...' : 'Refresh' }}
          </button>
        </div>
        <p *ngIf="error" class="error">{{ error }}</p>
        <p *ngIf="!error && !loading && appointments.length === 0" class="muted">
          No assigned appointments.
        </p>
        <ul *ngIf="appointments.length > 0">
          <li *ngFor="let appointment of appointments">
            <div class="top-row">
              <strong>{{ appointment.scheduled_at | date: 'medium' }}</strong>
              <span class="badge" [class]="'badge ' + appointment.status">
                {{ appointment.status | titlecase }}
              </span>
            </div>
            <div class="meta">Patient: {{ appointment.patient_id }}</div>
            <div class="reason">{{ appointment.reason }}</div>
            <div class="actions" *ngIf="appointment.status === 'pending'">
              <button type="button" (click)="updateStatus(appointment.id, 'approved')" [disabled]="processingId === appointment.id">
                Approve
              </button>
              <button type="button" (click)="updateStatus(appointment.id, 'rejected')" [disabled]="processingId === appointment.id">
                Reject
              </button>
            </div>
          </li>
        </ul>
      </div>
    </section>
  `,
  styles: [`
    .appointments { max-width: 900px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #a7b0be; margin-top: -0.4rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .list-header { display: flex; justify-content: space-between; align-items: center; }
    button { width: fit-content; padding: 0.55rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.7; cursor: not-allowed; }
    ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.75rem; }
    li { border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 8px; padding: 0.75rem; }
    .top-row { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
    .badge { border-radius: 999px; padding: 0.15rem 0.55rem; font-size: 0.78rem; border: 1px solid transparent; }
    .pending { color: #facc15; border-color: rgba(250, 204, 21, 0.5); }
    .approved { color: #22c55e; border-color: rgba(34, 197, 94, 0.5); }
    .rejected { color: #f87171; border-color: rgba(248, 113, 113, 0.5); }
    .meta { color: #94a3b8; margin-top: 0.25rem; font-size: 0.85rem; }
    .reason { margin-top: 0.4rem; }
    .actions { display: flex; gap: 0.5rem; margin-top: 0.75rem; }
    .error { color: #f87171; }
    .muted { color: #94a3b8; }
  `]
})
export class DoctorAppointmentsComponent implements OnInit {
  appointments: Appointment[] = [];
  loading = false;
  error = '';
  processingId = '';

  constructor(private appointmentsService: AppointmentsService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.appointmentsService.listDoctor()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (res) => {
          this.appointments = (res.appointments ?? []).sort(
            (a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime()
          );
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to load appointments';
        }
      });
  }

  updateStatus(id: string, status: 'approved' | 'rejected'): void {
    this.processingId = id;
    this.appointmentsService.updateStatus(id, status)
      .pipe(finalize(() => (this.processingId = '')))
      .subscribe({
        next: () => this.load(),
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to update appointment status';
        }
      });
  }
}

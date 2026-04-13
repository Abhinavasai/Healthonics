import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { Appointment, AppointmentsService } from '../../services/appointments.service';

@Component({
  selector: 'app-doctor-appointments',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe, TitleCasePipe],
  template: `
    <section class="appointments" data-cy="doctor-appointments-page">
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
        <div class="summary" *ngIf="appointments.length > 0">
          <button type="button" class="chip" (click)="activeFilter = 'all'" [class.active]="activeFilter === 'all'" data-cy="doctor-filter-all">All {{ appointments.length }}</button>
          <button type="button" class="chip" (click)="activeFilter = 'pending'" [class.active]="activeFilter === 'pending'" data-cy="doctor-filter-pending">Pending {{ pendingCount }}</button>
          <button type="button" class="chip" (click)="activeFilter = 'approved'" [class.active]="activeFilter === 'approved'" data-cy="doctor-filter-approved">Approved {{ approvedCount }}</button>
          <button type="button" class="chip" (click)="activeFilter = 'rejected'" [class.active]="activeFilter === 'rejected'" data-cy="doctor-filter-rejected">Rejected {{ rejectedCount }}</button>
        </div>
        <ul *ngIf="visibleAppointments.length > 0">
          <li *ngFor="let appointment of visibleAppointments" data-cy="doctor-appointment-item">
            <div class="row">
              <a class="row-link" [routerLink]="['/doctor/appointments', appointment.id]">
                <div class="top-row">
                  <strong>{{ appointment.scheduled_at | date: 'medium' }}</strong>
                  <span class="badge" [class]="'badge ' + appointment.status">
                    {{ appointment.status | titlecase }}
                  </span>
                </div>
                <div class="meta">Patient: {{ shortId(appointment.patient_id) }}</div>
                <div class="reason">{{ appointment.reason }}</div>
                <span class="hint">View details</span>
              </a>
              <div class="actions" *ngIf="appointment.status === 'pending'">
                <button type="button" (click)="updateStatus(appointment.id, 'approved')" [disabled]="processingId === appointment.id">
                  Approve
                </button>
                <button type="button" (click)="updateStatus(appointment.id, 'rejected')" [disabled]="processingId === appointment.id">
                  Reject
                </button>
              </div>
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
    .summary { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 0.75rem; }
    .chip { border-color: #334155; color: #cbd5e1; }
    .chip.active { border-color: #22d3ee; color: #22d3ee; }
    button { width: fit-content; padding: 0.55rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.7; cursor: not-allowed; }
    ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.75rem; }
    li { border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 8px; padding: 0; overflow: hidden; }
    .row { display: grid; gap: 0.65rem; padding: 0.75rem; }
    .row-link {
      display: grid;
      gap: 0.35rem;
      color: inherit;
      text-decoration: none;
    }
    .row-link:hover { background: rgba(34, 211, 238, 0.06); border-radius: 8px; margin: -0.35rem; padding: 0.35rem; }
    .hint { font-size: 0.78rem; color: #22d3ee; margin-top: 0.25rem; }
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
  activeFilter: 'all' | 'pending' | 'approved' | 'rejected' = 'all';
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

  get pendingCount(): number {
    return this.appointments.filter((a) => a.status === 'pending').length;
  }
  get approvedCount(): number {
    return this.appointments.filter((a) => a.status === 'approved').length;
  }
  get rejectedCount(): number {
    return this.appointments.filter((a) => a.status === 'rejected').length;
  }
  get visibleAppointments(): Appointment[] {
    if (this.activeFilter === 'all') {
      return this.appointments;
    }
    return this.appointments.filter((a) => a.status === this.activeFilter);
  }
  shortId(value: string): string {
    return value?.slice(0, 8) ?? '';
  }
}

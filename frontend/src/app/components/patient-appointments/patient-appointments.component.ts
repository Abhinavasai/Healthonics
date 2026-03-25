import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { Appointment, AppointmentsService, DoctorOption } from '../../services/appointments.service';

@Component({
  selector: 'app-patient-appointments',
  standalone: true,
  imports: [CommonModule, FormsModule, DatePipe, TitleCasePipe],
  template: `
    <section class="appointments">
      <h1>Patient Appointments</h1>
      <p class="subtitle">Request an appointment and track its current status.</p>

      <form class="card form" (ngSubmit)="submit()" #appointmentForm="ngForm">
        <h2>Request Appointment</h2>
        <label>
          Doctor
          <select
            name="doctorId"
            [(ngModel)]="selectedDoctorId"
            required
          >
            <option value="" disabled>Select a doctor</option>
            <option *ngFor="let doctor of doctors" [value]="doctor.id">
              {{ doctor.email }}
            </option>
          </select>
        </label>
        <label>
          Date and Time
          <input
            type="datetime-local"
            name="scheduledAt"
            [(ngModel)]="scheduledAt"
            required
          />
        </label>
        <label>
          Reason
          <textarea
            name="reason"
            [(ngModel)]="reason"
            required
            minlength="5"
            rows="3"
            placeholder="Brief reason for the appointment"
          ></textarea>
        </label>
        <button type="submit" [disabled]="appointmentForm.invalid || submitting || doctorsLoading || doctors.length === 0">
          {{ submitting ? 'Submitting...' : 'Submit Request' }}
        </button>
        <p *ngIf="doctorsLoading" class="muted">Loading doctors...</p>
        <p *ngIf="!doctorsLoading && doctors.length === 0" class="error">No doctors available right now.</p>
        <p *ngIf="formMessage" class="message">{{ formMessage }}</p>
      </form>

      <div class="card list">
        <div class="list-header">
          <h2>My Requests</h2>
          <button type="button" (click)="load()" [disabled]="loading">
            {{ loading ? 'Refreshing...' : 'Refresh' }}
          </button>
        </div>
        <div class="summary" *ngIf="appointments.length > 0">
          <span class="badge pending">Pending {{ pendingCount }}</span>
          <span class="badge approved">Approved {{ approvedCount }}</span>
          <span class="badge rejected">Rejected {{ rejectedCount }}</span>
        </div>
        <p *ngIf="error" class="error">{{ error }}</p>
        <p *ngIf="!error && !loading && appointments.length === 0" class="muted">
          No appointments yet.
        </p>
        <ul *ngIf="appointments.length > 0">
          <li *ngFor="let appointment of sortedAppointments">
            <div class="top-row">
              <strong>{{ appointment.scheduled_at | date: 'medium' }}</strong>
              <span class="badge" [class]="'badge ' + appointment.status">
                {{ appointment.status | titlecase }}
              </span>
            </div>
            <div class="meta">Doctor: {{ appointment.doctor_id }}</div>
            <div class="reason">{{ appointment.reason }}</div>
          </li>
        </ul>
      </div>
    </section>
  `,
  styles: [`
    .appointments { max-width: 900px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #a7b0be; margin-top: -0.4rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .form { display: grid; gap: 0.8rem; }
    label { display: grid; gap: 0.35rem; font-size: 0.9rem; }
    input, select, textarea { background: #0b1220; color: #e5e7eb; border: 1px solid #334155; border-radius: 8px; padding: 0.55rem; }
    button { width: fit-content; padding: 0.55rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.7; cursor: not-allowed; }
    .list-header { display: flex; justify-content: space-between; align-items: center; }
    .summary { display: flex; gap: 0.5rem; margin-bottom: 0.75rem; }
    ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.75rem; }
    li { border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 8px; padding: 0.75rem; }
    .top-row { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
    .badge { border-radius: 999px; padding: 0.15rem 0.55rem; font-size: 0.78rem; border: 1px solid transparent; }
    .pending { color: #facc15; border-color: rgba(250, 204, 21, 0.5); }
    .approved { color: #22c55e; border-color: rgba(34, 197, 94, 0.5); }
    .rejected { color: #f87171; border-color: rgba(248, 113, 113, 0.5); }
    .meta { color: #94a3b8; margin-top: 0.25rem; font-size: 0.85rem; }
    .reason { margin-top: 0.4rem; }
    .error { color: #f87171; }
    .message { color: #22d3ee; margin: 0; }
    .muted { color: #94a3b8; }
  `]
})
export class PatientAppointmentsComponent implements OnInit {
  selectedDoctorId = '';
  scheduledAt = '';
  reason = '';
  appointments: Appointment[] = [];
  doctors: DoctorOption[] = [];

  loading = false;
  doctorsLoading = false;
  submitting = false;
  error = '';
  formMessage = '';

  constructor(private appointmentsService: AppointmentsService) {}

  ngOnInit(): void {
    this.loadDoctors();
    this.load();
  }

  loadDoctors(): void {
    this.doctorsLoading = true;
    this.appointmentsService.listDoctors()
      .pipe(finalize(() => (this.doctorsLoading = false)))
      .subscribe({
        next: (res) => {
          this.doctors = res.doctors ?? [];
          if (!this.selectedDoctorId && this.doctors.length > 0) {
            this.selectedDoctorId = this.doctors[0].id;
          }
        },
        error: (err) => {
          this.formMessage = err?.error?.error ?? 'Unable to load doctors';
        }
      });
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.appointmentsService.listPatient()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (res) => {
          this.appointments = (res.appointments ?? []).sort(
            (a, b) => new Date(b.scheduled_at).getTime() - new Date(a.scheduled_at).getTime()
          );
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to load appointments';
        }
      });
  }

  get sortedAppointments(): Appointment[] {
    const rank = { pending: 0, approved: 1, rejected: 2 };
    return [...this.appointments].sort((a, b) => {
      const rankDelta = rank[a.status] - rank[b.status];
      if (rankDelta !== 0) {
        return rankDelta;
      }
      return new Date(b.scheduled_at).getTime() - new Date(a.scheduled_at).getTime();
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

  submit(): void {
    this.submitting = true;
    this.formMessage = '';
    const payload = {
      doctor_id: this.selectedDoctorId,
      scheduled_at: new Date(this.scheduledAt).toISOString(),
      reason: this.reason.trim()
    };
    this.appointmentsService.create(payload)
      .pipe(finalize(() => (this.submitting = false)))
      .subscribe({
        next: () => {
          this.formMessage = 'Appointment request submitted.';
          this.reason = '';
          this.scheduledAt = '';
          this.load();
        },
        error: (err) => {
          this.formMessage = err?.error?.error ?? 'Unable to submit request';
        }
      });
  }
}

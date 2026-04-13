import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { Appointment, AppointmentsService, DoctorOption } from '../../services/appointments.service';

@Component({
  selector: 'app-patient-appointments',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, DatePipe, TitleCasePipe],
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
        <div class="schedule-grid">
          <label>
            Date
            <div class="date-row">
              <input
                #dateInput
                type="date"
                name="selectedDate"
                [(ngModel)]="selectedDate"
                required
                (click)="openDatePicker()"
              />
              <button
                type="button"
                class="calendar-trigger"
                (click)="openDatePicker()"
                aria-label="Open calendar to pick a date"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <rect x="3" y="4" width="18" height="18" rx="2" />
                  <path d="M16 2v4M8 2v4M3 10h18" />
                </svg>
              </button>
            </div>
            <span class="field-hint">Tap the date field or calendar button to open the picker.</span>
          </label>
        </div>

        <div class="time-block card-sub">
          <h3 class="time-title">Time</h3>
          <label class="time-select-label">
            <span class="sr-only">Appointment time</span>
            <select
              name="selectedTime"
              [(ngModel)]="selectedTime"
              required
            >
              <option *ngFor="let slot of daySlots" [value]="slot.value">
                {{ slot.label }}
              </option>
            </select>
          </label>
        </div>

        <label>
          Reason
          <textarea
            name="reason"
            [(ngModel)]="reason"
            required
            minlength="1"
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
        <ul *ngIf="appointments.length > 0" data-cy="patient-appointments-list">
          <li *ngFor="let appointment of sortedAppointments">
            <div class="row-wrap">
              <a class="row-link" [routerLink]="['/patient/appointments', appointment.id]">
                <div class="top-row">
                  <strong>{{ appointment.scheduled_at | date: 'medium' }}</strong>
                  <span class="badge" [class]="'badge ' + appointment.status">
                    {{ appointment.status | titlecase }}
                  </span>
                </div>
                <div class="meta">Doctor: {{ appointment.doctor_id }}</div>
                <div class="reason">{{ appointment.reason }}</div>
                <span class="hint">View details</span>
              </a>
              <button
                *ngIf="appointment.status === 'pending' || appointment.status === 'approved'"
                type="button"
                class="cancel-btn"
                [disabled]="cancellingId === appointment.id"
                (click)="cancelAppointment(appointment.id, $event)"
                data-cy="patient-appointment-cancel"
              >
                Cancel
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
    .card-sub { background: rgba(11, 18, 32, 0.6); border: 1px solid rgba(148, 163, 184, 0.15); border-radius: 10px; padding: 0.8rem; }
    .form { display: grid; gap: 0.8rem; }
    label { display: grid; gap: 0.35rem; font-size: 0.9rem; }
    input, select, textarea { background: #0b1220; color: #e5e7eb; border: 1px solid #334155; border-radius: 8px; padding: 0.55rem; }
    button { width: fit-content; padding: 0.55rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.7; cursor: not-allowed; }
    .list-header { display: flex; justify-content: space-between; align-items: center; }
    .summary { display: flex; gap: 0.5rem; margin-bottom: 0.75rem; }
    ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.75rem; }
    li { border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 8px; padding: 0; overflow: hidden; }
    .row-wrap { display: flex; align-items: stretch; gap: 0.5rem; }
    .cancel-btn {
      align-self: center;
      margin: 0.5rem;
      white-space: nowrap;
      border-color: #f87171;
      color: #f87171;
    }
    .row-link {
      display: grid;
      gap: 0.35rem;
      padding: 0.75rem;
      color: inherit;
      text-decoration: none;
    }
    .row-link:hover { background: rgba(34, 211, 238, 0.06); }
    .hint { font-size: 0.78rem; color: #22d3ee; margin-top: 0.25rem; }
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
    .schedule-grid { display: grid; grid-template-columns: 1fr; gap: 0.75rem; }
    .date-row {
      display: flex;
      align-items: stretch;
      gap: 0.5rem;
    }
    .date-row input[type="date"] {
      flex: 1;
      min-width: 0;
      min-height: 44px;
    }
    .calendar-trigger {
      flex-shrink: 0;
      width: 44px;
      min-height: 44px;
      padding: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 8px;
    }
    .field-hint { color: #94a3b8; font-size: 0.78rem; margin-top: 0.25rem; }
    .time-block { display: grid; gap: 0.6rem; }
    .time-title { margin: 0; font-size: 0.95rem; color: #93c5fd; font-weight: 600; }
    .time-select-label select { width: 100%; max-width: 100%; }
    .sr-only {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip: rect(0, 0, 0, 0);
      white-space: nowrap;
      border: 0;
    }
  `]
})
export class PatientAppointmentsComponent implements OnInit {
  cancellingId: string | null = null;
  @ViewChild('dateInput') private dateInputRef?: ElementRef<HTMLInputElement>;

  selectedDoctorId = '';
  selectedDate = this.getTodayLocalDate();
  selectedTime = '09:00';
  reason = '';
  appointments: Appointment[] = [];
  doctors: DoctorOption[] = [];

  loading = false;
  doctorsLoading = false;
  submitting = false;
  error = '';
  formMessage = '';

  constructor(private appointmentsService: AppointmentsService) {}

  /** Opens native date picker (showPicker when supported; focus/click fallback). */
  openDatePicker(): void {
    const el = this.dateInputRef?.nativeElement;
    if (!el) {
      return;
    }
    const withPicker = el as HTMLInputElement & { showPicker?: () => void };
    if (typeof withPicker.showPicker === 'function') {
      try {
        withPicker.showPicker();
        return;
      } catch {
        // NotAllowedError or unsupported context — fall back
      }
    }
    el.focus();
    el.click();
  }

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
    const rank: Record<string, number> = {
      pending: 0,
      approved: 1,
      reschedule_requested: 1,
      rejected: 2,
      cancelled: 3,
      completed: 4,
      no_show: 4
    };
    return [...this.appointments].sort((a, b) => {
      const rankDelta = (rank[a.status] ?? 9) - (rank[b.status] ?? 9);
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

  get daySlots(): { value: string; label: string; minutes: number }[] {
    const slots: { value: string; label: string; minutes: number }[] = [];
    const startMinutes = 8 * 60;
    const endMinutes = 18 * 60;
    for (let m = startMinutes; m < endMinutes; m += 15) {
      const from = this.toHHMM(m);
      const to = this.toHHMM(m + 15);
      slots.push({
        value: from,
        label: `${this.toReadableTime(from)} - ${this.toReadableTime(to)}`,
        minutes: m
      });
    }
    return slots;
  }

  cancelAppointment(id: string, ev: Event): void {
    ev.preventDefault();
    ev.stopPropagation();
    this.cancellingId = id;
    this.appointmentsService
      .cancelAsPatient(id)
      .pipe(finalize(() => (this.cancellingId = null)))
      .subscribe({
        next: () => this.load(),
        error: (err) => {
          this.error = err?.error?.error ?? 'Could not cancel';
        }
      });
  }

  submit(): void {
    this.submitting = true;
    this.formMessage = '';
    const scheduledAt = this.combineDateAndTime(this.selectedDate, this.selectedTime);
    if (!scheduledAt) {
      this.submitting = false;
      this.formMessage = 'Please choose a valid date and 15-minute time slot.';
      return;
    }
    const payload = {
      doctor_id: this.selectedDoctorId,
      scheduled_at: scheduledAt,
      reason: this.reason.trim()
    };
    this.appointmentsService.create(payload)
      .pipe(finalize(() => (this.submitting = false)))
      .subscribe({
        next: () => {
          this.formMessage = 'Appointment request submitted.';
          this.reason = '';
          this.selectedDate = this.getTodayLocalDate();
          this.selectedTime = '09:00';
          this.load();
        },
        error: (err) => {
          this.formMessage = err?.error?.error ?? 'Unable to submit request';
        }
      });
  }

  private combineDateAndTime(date: string, time: string): string | null {
    const mins = this.parseHHMM(time);
    if (!date || mins === null || mins % 15 !== 0) {
      return null;
    }
    const [year, month, day] = date.split('-').map((v) => Number(v));
    const hours = Math.floor(mins / 60);
    const minutes = mins % 60;
    const local = new Date(year, (month ?? 1) - 1, day ?? 1, hours, minutes, 0, 0);
    if (Number.isNaN(local.getTime())) {
      return null;
    }
    return local.toISOString();
  }

  private parseHHMM(value: string): number | null {
    const trimmed = (value || '').trim();
    const match = /^([01]\d|2[0-3]):([0-5]\d)$/.exec(trimmed);
    if (!match) {
      return null;
    }
    return Number(match[1]) * 60 + Number(match[2]);
  }

  private toHHMM(totalMinutes: number): string {
    const normalized = ((totalMinutes % (24 * 60)) + (24 * 60)) % (24 * 60);
    const hours = Math.floor(normalized / 60).toString().padStart(2, '0');
    const minutes = (normalized % 60).toString().padStart(2, '0');
    return `${hours}:${minutes}`;
  }

  toReadableTime(hhmm: string): string {
    const mins = this.parseHHMM(hhmm);
    if (mins === null) {
      return hhmm;
    }
    const hour24 = Math.floor(mins / 60);
    const minute = mins % 60;
    const hour12 = hour24 % 12 === 0 ? 12 : hour24 % 12;
    const ampm = hour24 >= 12 ? 'PM' : 'AM';
    return `${hour12}:${minute.toString().padStart(2, '0')} ${ampm}`;
  }

  private getTodayLocalDate(): string {
    const now = new Date();
    const offset = now.getTimezoneOffset() * 60000;
    return new Date(now.getTime() - offset).toISOString().slice(0, 10);
  }
}

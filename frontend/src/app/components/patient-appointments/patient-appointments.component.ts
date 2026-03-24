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
        <div class="schedule-grid">
          <label>
            Date
            <input
              type="date"
              name="selectedDate"
              [(ngModel)]="selectedDate"
              required
            />
          </label>
          <label>
            Time (type or pick)
            <input
              type="time"
              step="900"
              name="selectedTime"
              [(ngModel)]="selectedTime"
              (ngModelChange)="onTimeTyped($event)"
              required
            />
          </label>
        </div>
        <div class="selected-range">
          Selected slot: <strong>{{ selectedRangeLabel }}</strong>
        </div>

        <div class="time-picker card-sub">
          <h3>Radial Selector</h3>
          <div class="radial">
            <button
              type="button"
              class="radial-slot"
              *ngFor="let slot of radialSlots; let i = index"
              [style.--i]="i"
              [style.--count]="radialSlots.length"
              [class.active]="slot.value === selectedTime"
              (click)="selectSlot(slot.value)"
            >
              {{ slot.value }}
            </button>
            <div class="radial-center">{{ toReadableTime(selectedTime) }}</div>
          </div>
        </div>

        <div class="time-slot-list">
          <button
            type="button"
            class="slot-chip"
            *ngFor="let slot of daySlots"
            [class.active]="slot.value === selectedTime"
            (click)="selectSlot(slot.value)"
          >
            {{ slot.label }}
          </button>
        </div>
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
    .card-sub { background: rgba(11, 18, 32, 0.6); border: 1px solid rgba(148, 163, 184, 0.15); border-radius: 10px; padding: 0.8rem; }
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
    .schedule-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
    .selected-range { color: #cbd5e1; font-size: 0.9rem; margin-top: -0.2rem; }
    .time-picker h3 { margin: 0 0 0.6rem; font-size: 0.95rem; color: #93c5fd; }
    .radial {
      position: relative;
      width: 270px;
      height: 270px;
      margin: 0.25rem auto;
      border-radius: 50%;
      border: 1px dashed rgba(148, 163, 184, 0.35);
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .radial-center {
      width: 90px;
      height: 90px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.8rem;
      text-align: center;
      border: 1px solid rgba(34, 211, 238, 0.6);
      color: #67e8f9;
      background: rgba(15, 23, 42, 0.9);
      padding: 0.35rem;
    }
    .radial-slot {
      position: absolute;
      width: 68px;
      height: 34px;
      padding: 0;
      border-radius: 999px;
      font-size: 0.75rem;
      line-height: 1;
      left: 50%;
      top: 50%;
      transform:
        translate(-50%, -50%)
        rotate(calc(var(--i) * (360deg / var(--count))))
        translateY(-112px)
        rotate(calc(-1 * var(--i) * (360deg / var(--count))));
      border: 1px solid rgba(148, 163, 184, 0.4);
      background: #0b1220;
      color: #cbd5e1;
      cursor: pointer;
    }
    .radial-slot.active {
      border-color: rgba(34, 211, 238, 0.85);
      color: #22d3ee;
      box-shadow: 0 0 0 1px rgba(34, 211, 238, 0.25);
    }
    .time-slot-list {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
      gap: 0.45rem;
      max-height: 180px;
      overflow: auto;
      padding-right: 0.25rem;
    }
    .slot-chip {
      width: 100%;
      text-align: left;
      font-size: 0.78rem;
      border: 1px solid rgba(148, 163, 184, 0.3);
      background: rgba(2, 6, 23, 0.9);
      color: #cbd5e1;
    }
    .slot-chip.active {
      border-color: rgba(34, 211, 238, 0.85);
      color: #22d3ee;
      background: rgba(8, 47, 73, 0.35);
    }
    @media (max-width: 700px) {
      .schedule-grid { grid-template-columns: 1fr; }
      .radial { width: 240px; height: 240px; }
      .radial-slot { transform:
        translate(-50%, -50%)
        rotate(calc(var(--i) * (360deg / var(--count))))
        translateY(-98px)
        rotate(calc(-1 * var(--i) * (360deg / var(--count))));
      }
    }
  `]
})
export class PatientAppointmentsComponent implements OnInit {
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

  get radialSlots(): { value: string; label: string; minutes: number }[] {
    const slots = this.daySlots;
    const selected = this.parseHHMM(this.selectedTime);
    if (selected === null) {
      return slots.slice(0, 12);
    }

    let centerIndex = slots.findIndex((s) => s.minutes === selected);
    if (centerIndex === -1) {
      centerIndex = slots.findIndex((s) => s.minutes > selected);
      centerIndex = centerIndex === -1 ? slots.length - 1 : centerIndex;
    }

    let start = Math.max(0, centerIndex - 5);
    if (start + 12 > slots.length) {
      start = Math.max(0, slots.length - 12);
    }
    return slots.slice(start, start + 12);
  }

  onTimeTyped(raw: string): void {
    const parsed = this.parseHHMM(raw);
    if (parsed === null) {
      return;
    }
    const snapped = Math.round(parsed / 15) * 15;
    this.selectedTime = this.toHHMM(snapped);
  }

  selectSlot(slotValue: string): void {
    this.selectedTime = slotValue;
  }

  get selectedRangeLabel(): string {
    const minutes = this.parseHHMM(this.selectedTime);
    if (minutes === null) {
      return 'Pick a valid time in 15-minute slots.';
    }
    return `${this.toReadableTime(this.toHHMM(minutes))} - ${this.toReadableTime(this.toHHMM(minutes + 15))}`;
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

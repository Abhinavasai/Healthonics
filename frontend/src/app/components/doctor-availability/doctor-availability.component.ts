import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';
import { GeoBookingService } from '../../services/geo-booking.service';
import { DoctorSlotRow } from '../../services/geo-booking.service';

/**
 * Doctor: create availability windows (POST /api/doctor/slots). List/delete open slots in a follow-up.
 */
@Component({
  selector: 'app-doctor-availability',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="avail">
      <h1>Availability</h1>
      <p class="subtitle">Add a slot using date + time (UTC). End time is computed from start time.</p>

      <div class="card grid" data-cy="doctor-availability-form">
        <label>
          Date (UTC)
          <input
            type="date"
            [(ngModel)]="selectedDate"
            data-cy="doctor-availability-date"
          />
        </label>

        <div class="times">
          <label>
            Start (UTC)
            <input
              type="time"
              step="300"
              [(ngModel)]="startTime"
              (ngModelChange)="onStartTimeChange($event)"
              data-cy="doctor-availability-start-time"
            />
          </label>
          <label>
            End (UTC)
            <input type="time" step="300" [(ngModel)]="endTime" data-cy="doctor-availability-end-time" />
          </label>
        </div>
      </div>

      <button type="button" (click)="submit()" [disabled]="saving" data-cy="doctor-availability-add-slot">
        {{ saving ? 'Adding...' : 'Add slot' }}
      </button>
      <p *ngIf="message" class="msg">{{ message }}</p>
      <p *ngIf="error" class="error">{{ error }}</p>

      <section *ngIf="slotsLoaded" class="slots" data-cy="doctor-availability-slots">
        <h2>Your slots</h2>
        <p *ngIf="slots.length === 0 && !error" class="muted" data-cy="doctor-availability-slots-empty">
          No slots yet.
        </p>

        <table *ngIf="slots.length > 0" class="slots-table" data-cy="doctor-availability-slots-table">
          <thead>
            <tr>
              <th>Start</th>
              <th>End</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr
              *ngFor="let s of slots"
              data-cy="doctor-slot-row"
              [attr.data-start-at]="s.start_at"
              [attr.data-available]="s.available ? 'true' : 'false'"
            >
              <td>{{ formatIsoUtcShort(s.start_at) }}</td>
              <td>{{ formatIsoUtcShort(s.end_at) }}</td>
              <td>
                <span [class]="s.available ? 'pill pill-ready' : 'pill'" data-cy="doctor-slot-availability">
                  {{ s.available ? 'Available' : 'Booked' }}
                </span>
              </td>
              <td>
                <button
                  *ngIf="s.available"
                  type="button"
                  class="ghost"
                  (click)="deleteSlot(s.id)"
                  [disabled]="saving"
                  data-cy="doctor-slot-delete"
                >
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </section>
  `,
  styles: [`
    .avail { max-width: 560px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #94a3b8; margin-top: -0.25rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .grid { display: grid; gap: 0.75rem; }
    .times { display: grid; gap: 0.75rem; grid-template-columns: 1fr 1fr; }
    label { display: grid; gap: 0.35rem; font-size: 0.9rem; }
    input { background: #0b1220; color: #e5e7eb; border: 1px solid #334155; border-radius: 8px; padding: 0.5rem; }
    button { width: fit-content; padding: 0.5rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.6; }
    .msg { color: #22d3ee; }
    .error { color: #f87171; }
    .slots-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
    th, td { border-bottom: 1px solid rgba(148, 163, 184, 0.2); padding: 0.4rem 0; text-align: left; }
    .muted { color: #94a3b8; }
    .pill { padding: 0.15rem 0.5rem; border-radius: 999px; border: 1px solid rgba(148, 163, 184, 0.25); }
    .pill-ready { border-color: rgba(34, 211, 238, 0.55); color: #22d3ee; }
    .ghost { border: 1px solid rgba(148, 163, 184, 0.35); background: transparent; color: #94a3b8; }
  `]
})
export class DoctorAvailabilityComponent implements OnInit {
  selectedDate = '';
  startTime = '09:00';
  endTime = '09:30';

  saving = false;
  message = '';
  error = '';

  slotsLoaded = false;
  slots: DoctorSlotRow[] = [];
  private doctorId = '';

  constructor(private geo: GeoBookingService, private auth: AuthService) {}

  ngOnInit(): void {
    const u = this.auth.getUser();
    if (!u || u.role !== 'doctor') {
      this.error = 'Doctor role required';
      return;
    }
    this.doctorId = u.id;
    this.selectedDate = this.getTodayUtcDatePlusDays(1);
    this.onStartTimeChange(this.startTime);
    this.loadSlots();
  }

  private loadSlots(): void {
    this.slotsLoaded = false;
    this.geo.listDoctorSlots(this.doctorId).subscribe({
      next: (res) => {
        this.slots = res.slots ?? [];
        this.slotsLoaded = true;
      },
      error: () => {
        this.error = 'Unable to load doctor slots';
        this.slotsLoaded = true;
      }
    });
  }

  onStartTimeChange(value: string): void {
    this.startTime = (value || '').trim();
    const startMins = this.parseHHMM(this.startTime);
    if (startMins === null) {
      return;
    }
    const endMins = startMins + 30; // default availability slot length
    this.endTime = this.toHHMM(endMins);
  }

  submit(): void {
    this.message = '';
    this.error = '';
    this.saving = true;

    const startIso = this.buildUtcIso(this.selectedDate, this.startTime);
    const endIso = this.buildUtcIso(this.selectedDate, this.endTime);

    if (!startIso || !endIso) {
      this.error = 'Choose a valid UTC date and time.';
      this.saving = false;
      return;
    }
    const start = new Date(startIso);
    const end = new Date(endIso);
    if (!(end.getTime() > start.getTime())) {
      this.error = 'End time must be after start time.';
      this.saving = false;
      return;
    }

    this.geo
      .createSlot({
        start_at: startIso,
        end_at: endIso
      })
      .subscribe({
        next: (res) => {
          this.message = `Slot created (${res.id}).`;
          this.saving = false;
          this.loadSlots();
        },
        error: (err) => {
          const raw = String(err?.error?.error ?? 'Could not create slot');
          this.error = raw.includes('already exists')
            ? 'A slot at this start time already exists. Choose another time.'
            : raw;
          this.saving = false;
        }
      });
  }

  deleteSlot(slotId: string): void {
    if (!slotId) return;
    this.message = '';
    this.error = '';
    this.saving = true;

    this.geo.deleteSlot(slotId).subscribe({
      next: () => {
        this.message = 'Slot deleted.';
        this.saving = false;
        this.loadSlots();
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Could not delete slot';
        this.saving = false;
      }
    });
  }

  private buildUtcIso(date: string, hhmm: string): string | null {
    if (!date) return null;
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(date.trim());
    if (!m) return null;

    const year = Number(m[1]);
    const month = Number(m[2]);
    const day = Number(m[3]);
    const mins = this.parseHHMM(hhmm);
    if (mins === null) return null;

    const hours = Math.floor(mins / 60);
    const minutes = mins % 60;

    const dt = new Date(Date.UTC(year, month - 1, day, hours, minutes, 0, 0));
    if (Number.isNaN(dt.getTime())) return null;
    return dt.toISOString();
  }

  private parseHHMM(value: string): number | null {
    const match = /^([01]\d|2[0-3]):([0-5]\d)$/.exec((value || '').trim());
    if (!match) return null;
    return Number(match[1]) * 60 + Number(match[2]);
  }

  private toHHMM(totalMinutes: number): string {
    const normalized = ((totalMinutes % (24 * 60)) + 24 * 60) % (24 * 60);
    const hours = Math.floor(normalized / 60).toString().padStart(2, '0');
    const minutes = (normalized % 60).toString().padStart(2, '0');
    return `${hours}:${minutes}`;
  }

  private getTodayUtcDatePlusDays(days: number): string {
    const now = new Date();
    const plus = new Date(now.getTime() + days * 24 * 60 * 60 * 1000);
    return plus.toISOString().slice(0, 10);
  }

  formatIsoUtcShort(iso: string): string {
    const dt = new Date(iso);
    if (Number.isNaN(dt.getTime())) return iso;
    const hh = dt.getUTCHours().toString().padStart(2, '0');
    const mm = dt.getUTCMinutes().toString().padStart(2, '0');
    return `${hh}:${mm} UTC`;
  }
}

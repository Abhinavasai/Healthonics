import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GeoBookingService } from '../../services/geo-booking.service';
import { AuthService } from '../../services/auth.service';

/**
 * Doctor: create availability windows (POST /api/doctor/slots). List/delete open slots in a follow-up.
 */
@Component({
  selector: 'app-doctor-availability',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="avail" data-cy="doctor-availability-page">
      <h1>Availability</h1>
      <p class="subtitle">Create and manage your open slots for patient booking.</p>

      <div class="card grid">
        <label>
          Start (local)
          <input type="datetime-local" [(ngModel)]="startAt" data-cy="doctor-slot-start" />
        </label>
        <label>
          End (local)
          <input type="datetime-local" [(ngModel)]="endAt" data-cy="doctor-slot-end" />
        </label>
      </div>

      <button type="button" (click)="submit()" [disabled]="saving" data-cy="doctor-slot-create">Add slot</button>
      <p *ngIf="message" class="msg" data-cy="doctor-slot-message">{{ message }}</p>
      <p *ngIf="error" class="error" data-cy="doctor-slot-error">{{ error }}</p>

      <section class="card list-wrap">
        <div class="list-head">
          <h2>Open slots</h2>
          <button type="button" class="small" (click)="loadSlots()" [disabled]="loadingSlots">Refresh</button>
        </div>

        <p *ngIf="loadingSlots" class="subtitle">Loading slots...</p>
        <p *ngIf="!loadingSlots && !slots.length" class="subtitle">No open slots yet.</p>

        <ul *ngIf="slots.length" class="slot-list">
          <li *ngFor="let slot of slots" data-cy="doctor-slot-item">
            <div>
              <strong>{{ slot.start_at | date: 'medium' }}</strong>
              <p>to {{ slot.end_at | date: 'medium' }}</p>
            </div>
            <button type="button" class="danger" (click)="remove(slot.id)" [disabled]="saving">Delete</button>
          </li>
        </ul>
      </section>
    </section>
  `,
  styles: [`
    .avail { max-width: 560px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #94a3b8; margin-top: -0.25rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .grid { display: grid; gap: 0.75rem; }
    label { display: grid; gap: 0.35rem; font-size: 0.9rem; }
    input { background: #0b1220; color: #e5e7eb; border: 1px solid #334155; border-radius: 8px; padding: 0.5rem; }
    button { width: fit-content; padding: 0.5rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.6; }
    .small { padding: 0.35rem 0.6rem; font-size: 0.85rem; }
    .danger { border-color: #f87171; color: #f87171; }
    .msg { color: #22d3ee; }
    .error { color: #f87171; }
    .list-wrap { margin-top: 0.5rem; }
    .list-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.25rem; }
    .list-head h2 { margin: 0; font-size: 1rem; }
    .slot-list { list-style: none; margin: 0; padding: 0; display: grid; gap: 0.5rem; }
    .slot-list li { border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 8px; padding: 0.65rem; display: flex; justify-content: space-between; gap: 0.75rem; }
    .slot-list p { margin: 0.2rem 0 0; color: #94a3b8; font-size: 0.85rem; }
  `]
})
export class DoctorAvailabilityComponent {
  startAt = '';
  endAt = '';
  saving = false;
  loadingSlots = false;
  message = '';
  error = '';
  slots: { id: string; start_at: string; end_at: string; available: boolean }[] = [];

  constructor(
    private geo: GeoBookingService,
    private auth: AuthService
  ) {}

  ngOnInit(): void {
    this.loadSlots();
  }

  submit(): void {
    this.message = '';
    this.error = '';
    this.saving = true;
    const start = new Date(this.startAt.trim());
    const end = new Date(this.endAt.trim());
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      this.error = 'Use valid ISO-8601 datetimes.';
      this.saving = false;
      return;
    }
    this.geo
      .createSlot({
        start_at: start.toISOString(),
        end_at: end.toISOString()
      })
      .subscribe({
        next: (res) => {
          this.message = `Slot created (${res.id}).`;
          this.startAt = '';
          this.endAt = '';
          this.loadSlots();
          this.saving = false;
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Could not create slot';
          this.saving = false;
        }
      });
  }

  loadSlots(): void {
    const doctorId = this.auth.getUser()?.id;
    if (!doctorId) {
      this.error = 'Could not detect doctor profile.';
      return;
    }
    this.loadingSlots = true;
    this.geo.listDoctorSlots(doctorId).subscribe({
      next: (res) => {
        this.slots = (res.slots ?? []).filter((s) => s.available);
        this.loadingSlots = false;
      },
      error: () => {
        this.error = 'Could not load slots.';
        this.loadingSlots = false;
      }
    });
  }

  remove(slotId: string): void {
    this.error = '';
    this.message = '';
    this.saving = true;
    this.geo.deleteSlot(slotId).subscribe({
      next: () => {
        this.message = 'Slot deleted.';
        this.loadSlots();
        this.saving = false;
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Could not delete slot';
        this.saving = false;
      }
    });
  }
}

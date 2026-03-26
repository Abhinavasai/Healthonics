import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GeoBookingService } from '../../services/geo-booking.service';

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
      <p class="subtitle">
        Add a slot using ISO-8601 times (UTC). Example: start 2026-03-27T14:00:00Z, end 2026-03-27T14:30:00Z.
      </p>

      <div class="card grid">
        <label>
          Start (ISO)
          <input type="text" [(ngModel)]="startAt" placeholder="2026-03-27T14:00:00Z" />
        </label>
        <label>
          End (ISO)
          <input type="text" [(ngModel)]="endAt" placeholder="2026-03-27T14:30:00Z" />
        </label>
      </div>

      <button type="button" (click)="submit()" [disabled]="saving">Add slot</button>
      <p *ngIf="message" class="msg">{{ message }}</p>
      <p *ngIf="error" class="error">{{ error }}</p>
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
    .msg { color: #22d3ee; }
    .error { color: #f87171; }
  `]
})
export class DoctorAvailabilityComponent {
  startAt = '';
  endAt = '';
  saving = false;
  message = '';
  error = '';

  constructor(private geo: GeoBookingService) {}

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
          this.saving = false;
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Could not create slot';
          this.saving = false;
        }
      });
  }
}

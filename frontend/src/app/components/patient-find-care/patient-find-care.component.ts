import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DoctorSearchRow, GeoBookingService, HospitalNear } from '../../services/geo-booking.service';

/**
 * Patient: location + radius + specialty search → hospitals & doctors → slot booking.
 * Scaffold: wire GeoBookingService; expand with geolocation, validation, and navigation.
 */
@Component({
  selector: 'app-patient-find-care',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="find-care">
      <h1>Find care near you</h1>
      <p class="subtitle">
        Set your search point (latitude / longitude), a radius in kilometers, and an optional specialty.
        Then load hospitals and matching doctors. Slot booking uses POST /api/appointments/book-slot (backend branch).
      </p>

      <div class="card grid">
        <label>
          Latitude
          <input type="number" step="any" [(ngModel)]="lat" placeholder="e.g. 29.64" />
        </label>
        <label>
          Longitude
          <input type="number" step="any" [(ngModel)]="lng" placeholder="e.g. -82.34" />
        </label>
        <label>
          Radius (km)
          <input type="number" step="1" min="1" [(ngModel)]="radiusKm" />
        </label>
        <label class="span-2">
          Specialty / problem keyword
          <input type="text" [(ngModel)]="specialization" placeholder="e.g. Internal Medicine" />
        </label>
      </div>

      <div class="actions">
        <button type="button" (click)="loadHospitals()" [disabled]="loading">Hospitals nearby</button>
        <button type="button" (click)="loadDoctors()" [disabled]="loading">Find doctors</button>
      </div>

      <p *ngIf="error" class="error">{{ error }}</p>

      <div *ngIf="hospitals.length" class="card">
        <h2>Hospitals</h2>
        <ul>
          <li *ngFor="let h of hospitals">
            {{ h.name }} — {{ h.city }}, {{ h.region }} ({{ h.distance_km }} km)
          </li>
        </ul>
      </div>

      <div *ngIf="doctors.length" class="card">
        <h2>Doctors</h2>
        <ul>
          <li *ngFor="let d of doctors">
            <strong>{{ d.email }}</strong>
            — {{ d.specialization || '—' }} ({{ d.distance_km }} km)
          </li>
        </ul>
      </div>

      <p class="muted">
        Next: geolocation, slot list per doctor, and book flow — see docs/geo-specialty-booking-frontend.md.
      </p>
    </section>
  `,
  styles: [`
    .find-care { max-width: 720px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #94a3b8; margin-top: -0.25rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
    .span-2 { grid-column: 1 / -1; }
    label { display: grid; gap: 0.35rem; font-size: 0.9rem; }
    input { background: #0b1220; color: #e5e7eb; border: 1px solid #334155; border-radius: 8px; padding: 0.5rem; }
    .actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
    button { padding: 0.5rem 0.85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
    button:disabled { opacity: 0.6; cursor: not-allowed; }
    ul { margin: 0; padding-left: 1.1rem; }
    .error { color: #f87171; }
    .muted { color: #64748b; font-size: 0.85rem; }
  `]
})
export class PatientFindCareComponent {
  lat = 29.64;
  lng = -82.34;
  radiusKm = 50;
  specialization = '';
  hospitals: HospitalNear[] = [];
  doctors: DoctorSearchRow[] = [];
  loading = false;
  error = '';

  constructor(private geo: GeoBookingService) {}

  loadHospitals(): void {
    this.error = '';
    this.loading = true;
    this.geo.hospitalsNear(this.lat, this.lng, this.radiusKm).subscribe({
      next: (res) => {
        this.hospitals = res.hospitals ?? [];
        this.loading = false;
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Could not load hospitals';
        this.loading = false;
      }
    });
  }

  loadDoctors(): void {
    this.error = '';
    this.loading = true;
    this.geo.searchDoctors(this.lat, this.lng, this.radiusKm, this.specialization).subscribe({
      next: (res) => {
        this.doctors = res.doctors ?? [];
        this.loading = false;
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Could not search doctors';
        this.loading = false;
      }
    });
  }
}

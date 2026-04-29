import { CommonModule } from '@angular/common';
import {
  AfterViewInit,
  Component,
  ElementRef,
  OnDestroy,
  ViewChild,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import * as L from 'leaflet';
import {
  DoctorSearchRow,
  GeocodeHit,
  GeoBookingService,
  HospitalNear,
} from '../../services/geo-booking.service';

/**
 * Patient: address/place search (OSM Nominatim via backend) + browser geolocation +
 * radius + specialty → hospitals & doctors. Map: Leaflet + OpenStreetMap tiles (free).
 * Swap geocode/map providers later without changing /api/hospitals/near contract.
 */
@Component({
  selector: 'app-patient-find-care',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="find-care">
      <h1>Find care near you</h1>
      <p class="subtitle">
        Search for a place or address, or use your current location. Then load nearby
        hospitals and matching doctors. Map uses OpenStreetMap (free tiles); search uses
        Nominatim through our API so we can switch to Google Maps in a later sprint.
      </p>

      <div class="card">
        <div class="search-row">
          <label class="grow">
            Location
            <input
              type="text"
              data-cy="location-query"
              [(ngModel)]="locationQuery"
              (keydown.enter)="runGeocodeSearch()"
              placeholder="City, address, landmark…"
              autocomplete="off"
            />
          </label>
          <button type="button" class="secondary" data-cy="location-search" (click)="runGeocodeSearch()" [disabled]="loadingGeocode">
            Search
          </button>
          <button type="button" class="secondary" data-cy="use-my-location" (click)="useMyLocation()" [disabled]="loadingGeo">
            Use my location
          </button>
        </div>
        <ul *ngIf="geocodeSuggestions.length" class="suggestions">
          <li *ngFor="let s of geocodeSuggestions">
            <button type="button" class="suggest-btn" data-cy="geocode-suggestion" (click)="pickGeocodeSuggestion(s)">
              {{ s.display_name }}
            </button>
          </li>
        </ul>
        <p *ngIf="selectedLocationLabel" class="selected-label">
          Search point: <strong>{{ selectedLocationLabel }}</strong>
          <span class="coords"> ({{ lat | number : '1.4-4' }}, {{ lng | number : '1.4-4' }})</span>
        </p>
      </div>

      <div class="card grid">
        <label>
          Radius (km)
          <input type="number" step="1" min="1" [(ngModel)]="radiusKm" />
        </label>
        <label>
          Department
          <select [(ngModel)]="department" (ngModelChange)="onDepartmentChange($event)">
            <option value="">All departments</option>
            <option *ngFor="let dep of departments" [value]="dep">{{ dep }}</option>
          </select>
        </label>
        <label class="span-2">
          Specialty / problem keyword
          <input type="text" [(ngModel)]="specialization" placeholder="e.g. Internal Medicine" />
        </label>
      </div>

      <div class="actions">
        <button type="button" data-cy="hospitals-nearby" (click)="loadHospitals()" [disabled]="loading">Hospitals nearby</button>
        <button type="button" data-cy="find-doctors" (click)="loadDoctors()" [disabled]="loading">Find doctors</button>
      </div>

      <div #mapHost class="map-host"></div>
      <p class="map-note">
        Tiles ©
        <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener">OpenStreetMap</a>
        contributors. Respect Nominatim usage limits when searching (prefer deliberate searches).
      </p>

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
    </section>
  `,
  styles: [
    `
      .find-care {
        max-width: 720px;
        margin: 0 auto;
        display: grid;
        gap: 1rem;
      }
      .subtitle {
        color: #94a3b8;
        margin-top: -0.25rem;
      }
      .card {
        background: rgba(15, 23, 42, 0.7);
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 12px;
        padding: 1rem;
      }
      .grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 0.75rem;
      }
      .span-2 {
        grid-column: 1 / -1;
      }
      .search-row {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        align-items: end;
      }
      .grow {
        flex: 1 1 200px;
      }
      label {
        display: grid;
        gap: 0.35rem;
        font-size: 0.9rem;
      }
      input,
      select {
        background: #0b1220;
        color: #e5e7eb;
        border: 1px solid #334155;
        border-radius: 8px;
        padding: 0.5rem;
      }
      .actions {
        display: flex;
        gap: 0.5rem;
        flex-wrap: wrap;
      }
      button {
        padding: 0.5rem 0.85rem;
        border-radius: 8px;
        border: 1px solid #22d3ee;
        background: #0f172a;
        color: #22d3ee;
        cursor: pointer;
      }
      button.secondary {
        border-color: #64748b;
        color: #e2e8f0;
      }
      button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
      }
      .suggestions {
        list-style: none;
        margin: 0.5rem 0 0;
        padding: 0;
        border: 1px solid rgba(148, 163, 184, 0.25);
        border-radius: 8px;
        max-height: 200px;
        overflow: auto;
      }
      .suggestions li {
        border-bottom: 1px solid rgba(148, 163, 184, 0.15);
      }
      .suggestions li:last-child {
        border-bottom: none;
      }
      .suggest-btn {
        width: 100%;
        text-align: left;
        border-radius: 0;
        border: none;
        background: transparent;
        color: #e2e8f0;
        padding: 0.5rem 0.65rem;
      }
      .suggest-btn:hover {
        background: rgba(34, 211, 238, 0.08);
      }
      .selected-label {
        margin: 0.75rem 0 0;
        font-size: 0.9rem;
        color: #94a3b8;
      }
      .coords {
        color: #64748b;
        font-size: 0.85rem;
      }
      .map-host {
        height: 280px;
        width: 100%;
        border-radius: 12px;
        border: 1px solid rgba(148, 163, 184, 0.25);
        z-index: 0;
      }
      .map-note {
        margin: -0.25rem 0 0;
        font-size: 0.75rem;
        color: #64748b;
      }
      .map-note a {
        color: #22d3ee;
      }
      ul {
        margin: 0;
        padding-left: 1.1rem;
      }
      .error {
        color: #f87171;
      }
    `,
  ],
})
export class PatientFindCareComponent implements AfterViewInit, OnDestroy {
  @ViewChild('mapHost') mapHost!: ElementRef<HTMLElement>;

  /** Default center (Gainesville area) until user searches or shares location. */
  lat = 29.6516;
  lng = -82.3248;
  radiusKm = 50;
  department = '';
  specialization = '';
  locationQuery = '';
  selectedLocationLabel = '';
  geocodeSuggestions: GeocodeHit[] = [];
  hospitals: HospitalNear[] = [];
  doctors: DoctorSearchRow[] = [];
  loading = false;
  loadingGeocode = false;
  loadingGeo = false;
  error = '';

  private map?: L.Map;
  private userMarker?: L.Marker;
  private hospitalsLayer?: L.LayerGroup;
  readonly departments = [
    'Cardiology',
    'Dermatology',
    'Emergency medicine',
    'Endocrinology',
    'Family medicine',
    'Gastroenterology',
    'Internal medicine',
    'Neurology',
    'Oncology',
    'Orthopedics',
    'Pediatrics',
    'Psychiatry',
    'Pulmonology',
    'Radiology'
  ];

  constructor(private geo: GeoBookingService) {}

  ngAfterViewInit(): void {
    queueMicrotask(() => this.initMap());
  }

  ngOnDestroy(): void {
    this.map?.remove();
    this.map = undefined;
  }

  private initMap(): void {
    if (!this.mapHost?.nativeElement || this.map) {
      return;
    }
    delete (L.Icon.Default.prototype as unknown as { _getIconUrl?: string })._getIconUrl;
    L.Icon.Default.mergeOptions({
      iconRetinaUrl: 'leaflet-images/marker-icon-2x.png',
      iconUrl: 'leaflet-images/marker-icon.png',
      shadowUrl: 'leaflet-images/marker-shadow.png',
    });

    this.map = L.map(this.mapHost.nativeElement).setView([this.lat, this.lng], 10);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '',
    }).addTo(this.map);

    this.userMarker = L.marker([this.lat, this.lng]).addTo(this.map);
    this.hospitalsLayer = L.layerGroup().addTo(this.map);
    this.selectedLocationLabel = this.selectedLocationLabel || 'Default map center';
  }

  private ensureMapReady(): void {
    if (!this.map) {
      this.initMap();
    }
  }

  private updateSearchPoint(lat: number, lng: number, label: string): void {
    this.lat = lat;
    this.lng = lng;
    this.selectedLocationLabel = label;
    this.geocodeSuggestions = [];
    this.ensureMapReady();
    if (this.userMarker && this.map) {
      this.userMarker.setLatLng([lat, lng]);
      this.map.setView([lat, lng], 11);
    }
  }

  runGeocodeSearch(): void {
    const q = this.locationQuery.trim();
    this.error = '';
    if (!q) {
      this.error = 'Enter a place or address to search.';
      return;
    }
    this.loadingGeocode = true;
    this.geocodeSuggestions = [];
    this.geo.geocodeSearch(q, 5).subscribe({
      next: (res) => {
        this.loadingGeocode = false;
        const hits = res.results ?? [];
        if (!hits.length) {
          this.error = 'No locations found. Try a different search.';
          return;
        }
        if (hits.length === 1) {
          const h = hits[0];
          this.updateSearchPoint(h.lat, h.lng, h.display_name);
          return;
        }
        this.geocodeSuggestions = hits;
      },
      error: (err) => {
        this.loadingGeocode = false;
        this.error = err?.error?.error ?? 'Location search failed.';
      },
    });
  }

  pickGeocodeSuggestion(s: GeocodeHit): void {
    this.locationQuery = s.display_name.split(',').slice(0, 2).join(',').trim();
    this.updateSearchPoint(s.lat, s.lng, s.display_name);
    this.loadHospitals();
  }

  useMyLocation(): void {
    if (!navigator.geolocation) {
      this.error = 'Geolocation is not supported in this browser.';
      return;
    }
    this.error = '';
    this.loadingGeo = true;
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        this.loadingGeo = false;
        const lat = pos.coords.latitude;
        const lng = pos.coords.longitude;
        this.locationQuery = 'My location';
        this.updateSearchPoint(lat, lng, 'My current location');
        this.loadHospitals();
      },
      (err) => {
        this.loadingGeo = false;
        const msg =
          err.code === err.PERMISSION_DENIED
            ? 'Location permission denied. Enable location or search manually.'
            : 'Could not read your location.';
        this.error = msg;
      },
      { enableHighAccuracy: false, timeout: 12_000, maximumAge: 60_000 }
    );
  }

  loadHospitals(): void {
    this.error = '';
    this.loading = true;
    this.geo.hospitalsNear(this.lat, this.lng, this.radiusKm).subscribe({
      next: (res) => {
        this.hospitals = res.hospitals ?? [];
        this.loading = false;
        this.plotHospitals();
        if (!this.hospitals.length) {
          this.error = 'No hospitals within this radius.';
        }
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Could not load hospitals';
        this.loading = false;
      },
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
      },
    });
  }

  onDepartmentChange(value: string): void {
    if (!this.specialization.trim()) {
      this.specialization = value;
    }
  }

  private plotHospitals(): void {
    this.ensureMapReady();
    if (!this.map || !this.hospitalsLayer) {
      return;
    }
    this.hospitalsLayer.clearLayers();
    const bounds: L.LatLngExpression[] = [[this.lat, this.lng]];
    for (const h of this.hospitals) {
      const ll: L.LatLngExpression = [h.latitude, h.longitude];
      bounds.push(ll);
      L.marker(ll)
        .bindPopup(`<strong>${h.name}</strong><br>${h.distance_km} km`)
        .addTo(this.hospitalsLayer);
    }
    if (this.hospitals.length) {
      this.map.fitBounds(L.latLngBounds(bounds), { padding: [28, 28], maxZoom: 12 });
    }
  }
}

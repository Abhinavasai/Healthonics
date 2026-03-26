import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface HospitalNear {
  id: string;
  name: string;
  city: string;
  region: string;
  latitude: number;
  longitude: number;
  distance_km: number;
}

export interface DoctorSearchRow {
  id: string;
  email: string;
  specialization: string;
  hospital_id?: string;
  distance_km: number;
}

export interface DoctorSlotRow {
  id: string;
  doctor_id: string;
  start_at: string;
  end_at: string;
  available: boolean;
}

/** OpenStreetMap Nominatim hit via GET /api/geocode (same shape works if you swap to Google later). */
export interface GeocodeHit {
  lat: number;
  lng: number;
  display_name: string;
}

@Injectable({ providedIn: 'root' })
export class GeoBookingService {
  private readonly api = '/api';

  constructor(private http: HttpClient) {}

  /** Text/address search → coordinates (backend proxies Nominatim; ~1 req/s policy on public OSM). */
  geocodeSearch(query: string, limit = 5): Observable<{ results: GeocodeHit[] }> {
    const q = query?.trim() ?? '';
    let p = new HttpParams().set('q', q).set('limit', String(limit));
    return this.http.get<{ results: GeocodeHit[] }>(`${this.api}/geocode`, { params: p });
  }

  hospitalsNear(lat: number, lng: number, radiusKm: number): Observable<{ hospitals: HospitalNear[] }> {
    let p = new HttpParams().set('lat', String(lat)).set('lng', String(lng)).set('radius_km', String(radiusKm));
    return this.http.get<{ hospitals: HospitalNear[] }>(`${this.api}/hospitals/near`, { params: p });
  }

  searchDoctors(lat: number, lng: number, radiusKm: number, specialization: string): Observable<{ doctors: DoctorSearchRow[] }> {
    let p = new HttpParams().set('lat', String(lat)).set('lng', String(lng)).set('radius_km', String(radiusKm));
    const spec = specialization?.trim();
    if (spec) {
      p = p.set('specialization', spec);
    }
    return this.http.get<{ doctors: DoctorSearchRow[] }>(`${this.api}/doctors/search`, { params: p });
  }

  listDoctorSlots(doctorId: string): Observable<{ slots: DoctorSlotRow[] }> {
    return this.http.get<{ slots: DoctorSlotRow[] }>(`${this.api}/doctors/${doctorId}/slots`);
  }

  createSlot(body: { start_at: string; end_at: string }): Observable<{ id: string }> {
    return this.http.post<{ id: string }>(`${this.api}/doctor/slots`, body);
  }

  deleteSlot(slotId: string): Observable<{ ok: boolean }> {
    return this.http.delete<{ ok: boolean }>(`${this.api}/doctor/slots/${slotId}`);
  }

  bookSlot(slotId: string, reason: string): Observable<unknown> {
    return this.http.post(`${this.api}/appointments/book-slot`, { slot_id: slotId, reason });
  }
}

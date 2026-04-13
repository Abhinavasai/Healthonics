import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

export interface Appointment {
  id: string;
  patient_id: string;
  doctor_id: string;
  scheduled_at: string;
  reason: string;
  status: 'pending' | 'approved' | 'rejected';
  created_at: string;
  updated_at: string;
}

export interface AppointmentListResponse {
  appointments: Appointment[];
}

export interface DoctorOption {
  id: string;
  email: string;
}

export interface DoctorListResponse {
  doctors: DoctorOption[];
}

/** Audit row from GET /api/appointments/:id/activity */
export interface AppointmentActivity {
  id: string;
  appointment_id: string;
  actor_user_id: string;
  /** When provided by API (joined user email). */
  actor_email?: string;
  action: string;
  detail?: string;
  created_at: string;
}

interface AppointmentActivityListResponse {
  activities?: AppointmentActivity[];
}

@Injectable({ providedIn: 'root' })
export class AppointmentsService {
  private readonly API = '/api/appointments';
  private readonly CoreAPI = '/api';

  constructor(private http: HttpClient) {}

  create(input: { doctor_id: string; scheduled_at: string; reason: string }): Observable<Appointment> {
    return this.http.post<Appointment>(this.API, input);
  }

  getById(id: string): Observable<Appointment> {
    return this.http.get<Appointment>(`${this.API}/${id}`);
  }

  getActivity(id: string): Observable<AppointmentActivity[]> {
    return this.http
      .get<AppointmentActivityListResponse>(`${this.API}/${id}/activity`)
      .pipe(map((r) => r.activities ?? []));
  }

  listPatient(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/patient`);
  }

  listDoctor(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/doctor`);
  }

  listDoctors(): Observable<DoctorListResponse> {
    return this.http.get<DoctorListResponse>(`${this.CoreAPI}/doctors`);
  }

  updateStatus(id: string, status: 'approved' | 'rejected'): Observable<Appointment> {
    return this.http.patch<Appointment>(`${this.API}/${id}/status`, { status });
  }
}

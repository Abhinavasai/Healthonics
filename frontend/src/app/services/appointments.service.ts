import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

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

@Injectable({ providedIn: 'root' })
export class AppointmentsService {
  private readonly API = '/api/appointments';

  constructor(private http: HttpClient) {}

  create(input: { doctor_id: string; scheduled_at: string; reason: string }): Observable<Appointment> {
    return this.http.post<Appointment>(this.API, input);
  }

  listPatient(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/patient`);
  }

  listDoctor(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/doctor`);
  }

  updateStatus(id: string, status: 'approved' | 'rejected'): Observable<Appointment> {
    return this.http.patch<Appointment>(`${this.API}/${id}/status`, { status });
  }
}

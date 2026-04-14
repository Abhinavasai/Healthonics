import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Appointment {
  id: string;
  patient_id: string;
  doctor_id: string;
  scheduled_at: string;
  reason: string;
  status: string;
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

  listPatient(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/patient`);
  }

  listDoctor(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(`${this.API}/doctor`);
  }

  listDoctors(): Observable<DoctorListResponse> {
    return this.http.get<DoctorListResponse>(`${this.CoreAPI}/doctors`);
  }

  listComments(appointmentId: string): Observable<{ comments: { id: string; author_user_id: string; body: string; created_at: string }[] }> {
    return this.http.get<{ comments: { id: string; author_user_id: string; body: string; created_at: string }[] }>(
      `${this.CoreAPI}/appointments/${appointmentId}/comments`
    );
  }

  postComment(appointmentId: string, body: string): Observable<{ id: string }> {
    return this.http.post<{ id: string }>(`${this.CoreAPI}/appointments/${appointmentId}/comments`, { body });
  }

  updateStatus(id: string, status: string): Observable<Appointment> {
    return this.http.patch<Appointment>(`${this.API}/${id}/status`, { status });
  }

  cancelAsPatient(id: string): Observable<Appointment> {
    return this.http.post<Appointment>(`${this.CoreAPI}/patient/appointments/${id}/cancel`, {});
  }
}

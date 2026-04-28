import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

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

/** Aligns with backend PR-23: internal vs patient_visible appointment comments. */
export type CommentVisibility = 'internal' | 'patient_visible';

export interface AppointmentComment {
  id: string;
  author_user_id: string;
  body: string;
  created_at: string;
  visibility?: CommentVisibility;
}

@Injectable({ providedIn: 'root' })
export class AppointmentsService {
  constructor(private http: HttpClient) {}

  create(input: { doctor_id: string; scheduled_at: string; reason: string }): Observable<Appointment> {
    return this.http.post<Appointment>(ApiContract.appointments.base, input);
  }

  getById(id: string): Observable<Appointment> {
    return this.http.get<Appointment>(ApiContract.appointments.byId(id));
  }

  listPatient(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(ApiContract.appointments.listPatient);
  }

  listDoctor(): Observable<AppointmentListResponse> {
    return this.http.get<AppointmentListResponse>(ApiContract.appointments.listDoctor);
  }

  listDoctors(): Observable<DoctorListResponse> {
    return this.http.get<DoctorListResponse>(ApiContract.geoBooking.doctors);
  }

  listComments(appointmentId: string): Observable<{ comments: AppointmentComment[] }> {
    return this.http.get<{ comments: AppointmentComment[] }>(ApiContract.appointments.comments(appointmentId));
  }

  postComment(
    appointmentId: string,
    body: string,
    visibility?: CommentVisibility
  ): Observable<{ id: string; visibility?: CommentVisibility }> {
    const payload: { body: string; visibility?: CommentVisibility } = { body };
    if (visibility !== undefined) {
      payload.visibility = visibility;
    }
    return this.http.post<{ id: string; visibility?: CommentVisibility }>(
      ApiContract.appointments.comments(appointmentId),
      payload
    );
  }

  updateStatus(id: string, status: string): Observable<Appointment> {
    return this.http.patch<Appointment>(ApiContract.appointments.updateStatus(id), { status });
  }

  cancelAsPatient(id: string): Observable<Appointment> {
    return this.http.post<Appointment>(ApiContract.appointments.patientCancel(id), {});
  }
}

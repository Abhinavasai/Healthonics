import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PatientDashboardSummary {
  role: string;
  upcoming_or_active_count: number;
  pending_requests: number;
}

export interface DoctorDashboardSummary {
  role: string;
  appointments_today: number;
  pending_queue: number;
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  constructor(private http: HttpClient) {}

  patientSummary(): Observable<PatientDashboardSummary> {
    return this.http.get<PatientDashboardSummary>('/api/patient/dashboard/summary');
  }

  doctorSummary(): Observable<DoctorDashboardSummary> {
    return this.http.get<DoctorDashboardSummary>('/api/doctor/dashboard/summary');
  }
}

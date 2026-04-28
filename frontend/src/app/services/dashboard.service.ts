import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

/** Matches backend PR-25 dashboard alerts. */
export type DashboardAlertSeverity = 'info' | 'warning';

export interface DashboardAlert {
  severity: DashboardAlertSeverity;
  code: string;
  message: string;
  count?: number;
}

export interface PatientDashboardAggregations {
  unread_messages: number;
  active_prescriptions: number;
  documents_uploaded: number;
}

export interface PatientDashboardSummary {
  role: string;
  upcoming_or_active_count: number;
  pending_requests: number;
  aggregations?: PatientDashboardAggregations;
  alerts?: DashboardAlert[];
}

export interface DoctorDashboardAggregations {
  unread_messages: number;
  appointments_this_week: number;
}

export interface DoctorDashboardSummary {
  role: string;
  appointments_today: number;
  pending_queue: number;
  aggregations?: DoctorDashboardAggregations;
  alerts?: DashboardAlert[];
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  constructor(private http: HttpClient) {}

  patientSummary(): Observable<PatientDashboardSummary> {
    return this.http.get<PatientDashboardSummary>(ApiContract.dashboard.patientSummary);
  }

  doctorSummary(): Observable<DoctorDashboardSummary> {
    return this.http.get<DoctorDashboardSummary>(ApiContract.dashboard.doctorSummary);
  }
}

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
  critical_open_count?: number;
}

export interface DoctorCriticalEscalation {
  id: string;
  document_id: string;
  patient_id: string;
  patient_email: string;
  filename: string;
  tier: number;
  status: 'open' | 'acknowledged' | 'resolved';
  rule_code: string;
  message: string;
  first_detected_at: string;
  next_escalation_at: string;
  acknowledged_at?: string;
}

export interface DoctorDashboardSummary {
  role: string;
  appointments_today: number;
  pending_queue: number;
  aggregations?: DoctorDashboardAggregations;
  alerts?: DashboardAlert[];
  critical_escalations?: DoctorCriticalEscalation[];
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

  acknowledgeCriticalEscalation(id: string): Observable<{ id: string; status: string }> {
    return this.http.post<{ id: string; status: string }>(
      `${ApiContract.dashboard.doctorCriticalEscalations}/${id}/ack`,
      {}
    );
  }
}

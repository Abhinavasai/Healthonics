import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PrescriptionRow {
  id: string;
  patient_id: string;
  doctor_id: string;
  medication_name: string;
  dosage: string;
  frequency: string;
  duration_days: number;
  instructions: string;
  status: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class PrescriptionsService {
  private readonly API = '/api';

  constructor(private http: HttpClient) {}

  listByPatient(patientId: string): Observable<{ prescriptions: PrescriptionRow[] }> {
    return this.http.get<{ prescriptions: PrescriptionRow[] }>(
      `${this.API}/patients/${patientId}/prescriptions`
    );
  }
}

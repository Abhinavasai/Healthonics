import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';
import { Subscription } from 'rxjs';
import { ApiContract } from '../../services/api-contract';

interface DoctorPatient {
  patient_id: string;
  email: string;
  last_appointment_at: string | null;
  active_prescriptions: number;
}

@Component({
  selector: 'app-doctor-patient-panel',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <section class="panel">
      <h1>My Patients</h1>
      <p *ngIf="loading" class="muted">Loading...</p>
      <p *ngIf="error" class="error">{{ error }}</p>
      <table *ngIf="!loading && patients.length" class="pt">
        <thead>
          <tr>
            <th>Patient</th>
            <th>Last Appointment</th>
            <th>Active Rx</th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let p of patients">
            <td>{{ p.email }}</td>
            <td>{{ p.last_appointment_at ? (p.last_appointment_at | date:'mediumDate') : 'None' }}</td>
            <td class="rx-count">{{ p.active_prescriptions }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!loading && !patients.length" class="muted">No patients yet.</p>
    </section>
  `,
  styles: [`
    .panel { max-width: 900px; margin: 0 auto; padding: 1rem; }
    h1 { margin: 0 0 1rem; font-size: 1.4rem; }
    .muted { color: #94a3b8; }
    .error { color: #f87171; }
    .pt { width: 100%; border-collapse: collapse; }
    .pt th { text-align: left; padding: 0.6rem 0.75rem; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em; color: #64748b; border-bottom: 1px solid #1e293b; }
    .pt td { padding: 0.65rem 0.75rem; border-bottom: 1px solid rgba(30,41,59,0.6); color: #e2e8f0; font-size: 0.9rem; }
    .pt tr:hover td { background: rgba(255,255,255,0.02); }
    .rx-count { text-align: center; font-weight: 700; color: #60a5fa; }
  `]
})
export class DoctorPatientPanelComponent implements OnInit, OnDestroy {
  patients: DoctorPatient[] = [];
  loading = false;
  error = '';
  private sub?: Subscription;

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.loading = true;
    this.sub = this.http.get<{ patients: DoctorPatient[] }>(ApiContract.doctor.patients).subscribe({
      next: (res) => { this.patients = res.patients ?? []; this.loading = false; },
      error: () => { this.error = 'Could not load patients'; this.loading = false; }
    });
  }

  ngOnDestroy(): void { this.sub?.unsubscribe(); }
}

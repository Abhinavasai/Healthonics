import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { PrescriptionsService, PrescriptionRow } from '../../services/prescriptions.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-patient-prescriptions',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section data-cy="patient-prescriptions-page">
      <h1>My prescriptions</h1>
      <p class="err" *ngIf="error">{{ error }}</p>
      <ul *ngIf="rows.length" data-cy="patient-prescriptions-list">
        <li *ngFor="let r of rows" data-cy="patient-prescriptions-row">
          <strong>{{ r.medication_name }}</strong> — {{ r.dosage }}, {{ r.frequency }}
          <span class="st">{{ r.status }}</span>
        </li>
      </ul>
      <p *ngIf="!error && !rows.length" data-cy="patient-prescriptions-empty">No prescriptions yet.</p>
    </section>
  `,
  styles: [
    `
      .err {
        color: #b00020;
      }
      .st {
        margin-left: 0.5rem;
        text-transform: uppercase;
        font-size: 0.75rem;
      }
    `
  ]
})
export class PatientPrescriptionsComponent implements OnInit {
  rows: PrescriptionRow[] = [];
  error: string | null = null;

  constructor(
    private rx: PrescriptionsService,
    private auth: AuthService
  ) {}

  ngOnInit(): void {
    const u = this.auth.getUser();
    if (!u) {
      this.error = 'Not signed in';
      return;
    }
    this.rx.listByPatient(u.id).subscribe({
      next: (res) => (this.rows = res.prescriptions ?? []),
      error: () => (this.error = 'Could not load prescriptions')
    });
  }
}

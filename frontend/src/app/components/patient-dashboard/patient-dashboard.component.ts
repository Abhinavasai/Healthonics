import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DashboardService, PatientDashboardSummary } from '../../services/dashboard.service';

@Component({
  selector: 'app-patient-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section class="dash" data-cy="patient-dashboard">
      <h1>Dashboard</h1>
      <p *ngIf="error" class="err">{{ error }}</p>
      <div *ngIf="data" class="cards" data-cy="patient-dashboard-stats">
        <div class="card">
          <span class="num">{{ data.upcoming_or_active_count }}</span>
          <span class="lbl">Upcoming / active</span>
        </div>
        <div class="card">
          <span class="num">{{ data.pending_requests }}</span>
          <span class="lbl">Pending requests</span>
        </div>
      </div>
    </section>
  `,
  styles: [
    `
      .dash {
        max-width: 720px;
      }
      .cards {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
      }
      .card {
        border: 1px solid rgba(148, 163, 184, 0.3);
        border-radius: 10px;
        padding: 1rem;
        min-width: 140px;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
      }
      .num {
        font-size: 1.75rem;
        font-weight: 700;
      }
      .lbl {
        color: #94a3b8;
        font-size: 0.85rem;
      }
      .err {
        color: #f87171;
      }
    `
  ]
})
export class PatientDashboardComponent implements OnInit {
  data: PatientDashboardSummary | null = null;
  error: string | null = null;

  constructor(private dash: DashboardService) {}

  ngOnInit(): void {
    this.dash.patientSummary().subscribe({
      next: (d) => (this.data = d),
      error: () => (this.error = 'Could not load summary')
    });
  }
}

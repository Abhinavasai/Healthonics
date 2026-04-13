import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { DashboardService, PatientDashboard } from '../../services/dashboard.service';

@Component({
  selector: 'app-patient-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe],
  template: `
    <section class="dash">
      <h1>Dashboard</h1>
      <p class="sub">Overview of your care activity.</p>
      <p *ngIf="error" class="err">{{ error }}</p>
      <p *ngIf="loading" class="muted">Loading…</p>
      <div *ngIf="data && !loading" class="grid">
        <div class="card">
          <span class="label">Upcoming (scheduled)</span>
          <strong>{{ data.upcoming_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Pending approval</span>
          <strong>{{ data.pending_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Unread messages</span>
          <strong>{{ data.unread_messages }}</strong>
        </div>
      </div>
      <ng-container *ngIf="data?.next_appointment as next">
        <div class="card next">
          <h2>Next visit</h2>
          <p>
            <strong>{{ next.scheduled_at | date: 'medium' }}</strong>
            — {{ next.status }}
          </p>
          <p class="muted">Doctor: {{ next.doctor_email }}</p>
          <a [routerLink]="['/patient/appointments', next.id]">View details</a>
        </div>
      </ng-container>
      <div class="links">
        <a routerLink="/patient/find-care">Find care</a>
        <a routerLink="/patient/appointments">My appointments</a>
        <a routerLink="/patient/messages">Messages</a>
      </div>
    </section>
  `,
  styles: [
    `
      .dash {
        max-width: 720px;
        margin: 0 auto;
        display: grid;
        gap: 1rem;
      }
      .sub {
        color: #94a3b8;
        margin: -0.25rem 0 0;
      }
      .grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
        gap: 0.75rem;
      }
      .card {
        background: rgba(15, 23, 42, 0.75);
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 12px;
        padding: 1rem;
        display: grid;
        gap: 0.35rem;
      }
      .label {
        font-size: 0.8rem;
        color: #94a3b8;
      }
      .next h2 {
        margin: 0 0 0.5rem;
        font-size: 1rem;
      }
      .links {
        display: flex;
        flex-wrap: wrap;
        gap: 0.75rem;
      }
      .links a {
        color: #22d3ee;
      }
      .err {
        color: #f87171;
      }
      .muted {
        color: #94a3b8;
      }
    `,
  ],
})
export class PatientDashboardComponent implements OnInit {
  data: PatientDashboard | null = null;
  loading = false;
  error = '';

  constructor(private dash: DashboardService) {}

  ngOnInit(): void {
    this.loading = true;
    this.dash
      .patient()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (d) => (this.data = d),
        error: (err) => (this.error = err?.error?.error ?? 'Unable to load dashboard'),
      });
  }
}

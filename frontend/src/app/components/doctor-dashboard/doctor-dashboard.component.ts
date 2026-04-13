import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { DashboardService, DoctorDashboard } from '../../services/dashboard.service';

@Component({
  selector: 'app-doctor-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <section class="dash">
      <h1>Dashboard</h1>
      <p class="sub">Today’s load and pending work.</p>
      <p *ngIf="error" class="err">{{ error }}</p>
      <p *ngIf="loading" class="muted">Loading…</p>
      <div *ngIf="data && !loading" class="grid">
        <div class="card">
          <span class="label">Today’s appointments</span>
          <strong>{{ data.today_appointments_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Pending requests</span>
          <strong>{{ data.pending_queue_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Unread messages</span>
          <strong>{{ data.unread_messages }}</strong>
        </div>
      </div>
      <div class="links">
        <a routerLink="/doctor/appointments">Appointment queue</a>
        <a routerLink="/doctor/availability">Availability</a>
        <a routerLink="/doctor/messages">Messages</a>
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
export class DoctorDashboardComponent implements OnInit {
  data: DoctorDashboard | null = null;
  loading = false;
  error = '';

  constructor(private dash: DashboardService) {}

  ngOnInit(): void {
    this.loading = true;
    this.dash
      .doctor()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (d) => (this.data = d),
        error: (err) => (this.error = err?.error?.error ?? 'Unable to load dashboard'),
      });
  }
}

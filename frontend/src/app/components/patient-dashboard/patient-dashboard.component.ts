import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Subscription } from 'rxjs';
import { Appointment, AppointmentsService, DoctorOption } from '../../services/appointments.service';

@Component({
  selector: 'app-patient-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe],
  template: `
    <section class="dashboard" data-cy="patient-dashboard-page">
      <header class="dashboard__header">
        <h1>Patient dashboard</h1>
        <p>Your current care summary and next appointments.</p>
      </header>

      <div class="cards">
        <article class="card">
          <span class="card__label">Total appointments</span>
          <strong data-cy="patient-dashboard-total">{{ appointments.length }}</strong>
        </article>
        <article class="card">
          <span class="card__label">Pending</span>
          <strong data-cy="patient-dashboard-pending">{{ pendingCount }}</strong>
        </article>
        <article class="card">
          <span class="card__label">Approved</span>
          <strong data-cy="patient-dashboard-approved">{{ approvedCount }}</strong>
        </article>
      </div>

      <section class="panel">
        <div class="panel__head">
          <h2>Upcoming appointments</h2>
          <a routerLink="/patient/appointments">View all</a>
        </div>

        <p *ngIf="loading" class="muted">Loading dashboard...</p>
        <p *ngIf="error" class="error" data-cy="patient-dashboard-error">{{ error }}</p>
        <p *ngIf="!loading && !error && upcoming.length === 0" class="muted">No upcoming appointments yet.</p>

        <ul *ngIf="!loading && !error && upcoming.length > 0" class="list">
          <li *ngFor="let a of upcoming" data-cy="patient-dashboard-upcoming-item">
            <div>
              <strong>{{ doctorLabel(a) }}</strong>
              <p>{{ a.reason }}</p>
            </div>
            <time>{{ a.scheduled_at | date: 'medium' }}</time>
          </li>
        </ul>
      </section>
    </section>
  `,
  styles: [`
    .dashboard { max-width: 1000px; margin: 0 auto; display: grid; gap: 1rem; }
    .dashboard__header h1 { margin: 0; }
    .dashboard__header p { margin: .35rem 0 0; color: var(--text-muted); }
    .cards { display: grid; gap: .75rem; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); }
    .card { border: var(--border-subtle); border-radius: var(--radius-md); padding: .9rem; background: var(--gradient-card); }
    .card__label { display: block; color: var(--text-muted); font-size: .85rem; margin-bottom: .35rem; }
    .card strong { font-size: 1.5rem; }
    .panel { border: var(--border-subtle); border-radius: var(--radius-md); background: var(--gradient-card); padding: 1rem; }
    .panel__head { display: flex; justify-content: space-between; align-items: center; margin-bottom: .5rem; }
    .list { margin: 0; padding: 0; list-style: none; display: grid; gap: .5rem; }
    .list li { display: flex; justify-content: space-between; gap: 1rem; border: var(--border-subtle); border-radius: var(--radius-sm); padding: .65rem .75rem; }
    .list p { margin: .25rem 0 0; color: var(--text-muted); }
    .muted { color: var(--text-muted); }
    .error { color: #f87171; }
  `]
})
export class PatientDashboardComponent implements OnInit, OnDestroy {
  loading = true;
  error = '';
  appointments: Appointment[] = [];
  doctors: DoctorOption[] = [];

  private readonly sub = new Subscription();

  constructor(private appointmentsService: AppointmentsService) {}

  ngOnInit(): void {
    this.sub.add(this.appointmentsService.listDoctors().subscribe({
      next: (res) => {
        this.doctors = res.doctors ?? [];
      }
    }));

    this.sub.add(this.appointmentsService.listPatient().subscribe({
      next: (res) => {
        this.appointments = res.appointments ?? [];
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to load dashboard data.';
        this.loading = false;
      }
    }));
  }

  ngOnDestroy(): void {
    this.sub.unsubscribe();
  }

  get pendingCount(): number {
    return this.appointments.filter((a) => a.status === 'pending').length;
  }

  get approvedCount(): number {
    return this.appointments.filter((a) => a.status === 'approved').length;
  }

  get upcoming(): Appointment[] {
    const now = Date.now();
    return this.appointments
      .filter((a) => new Date(a.scheduled_at).getTime() >= now)
      .sort((a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime())
      .slice(0, 5);
  }

  doctorLabel(appointment: Appointment): string {
    const doctor = this.doctors.find((d) => d.id === appointment.doctor_id);
    return doctor?.email ?? appointment.doctor_id;
  }
}

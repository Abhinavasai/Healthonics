import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import {
  AdminDashboard,
  AdminUserRow,
  DashboardService,
} from '../../services/dashboard.service';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="admin">
      <h1>Admin</h1>
      <p class="sub">System overview and user accounts.</p>

      <p *ngIf="dashError" class="err">{{ dashError }}</p>
      <div *ngIf="dash" class="grid">
        <div class="card">
          <span class="label">Database</span>
          <strong [class.bad]="!dash.db_ok">{{ dash.db_ok ? 'OK' : 'Error' }}</strong>
        </div>
        <div class="card">
          <span class="label">Patients</span>
          <strong>{{ dash.patients_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Doctors</span>
          <strong>{{ dash.doctors_count }}</strong>
        </div>
        <div class="card">
          <span class="label">Admins</span>
          <strong>{{ dash.admins_count }}</strong>
        </div>
      </div>

      <h2>Users</h2>
      <p *ngIf="usersError" class="err">{{ usersError }}</p>
      <p *ngIf="usersLoading" class="muted">Loading users…</p>
      <table *ngIf="users.length" class="tbl">
        <thead>
          <tr>
            <th>Email</th>
            <th>Role</th>
            <th>Active</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let u of users">
            <td>{{ u.email }}</td>
            <td>{{ u.role }}</td>
            <td>{{ u.active ? 'Yes' : 'No' }}</td>
            <td>
              <button
                type="button"
                *ngIf="u.active"
                (click)="toggleActive(u, false)"
                [disabled]="processingId === u.id"
              >
                Disable
              </button>
              <button
                type="button"
                *ngIf="!u.active"
                (click)="toggleActive(u, true)"
                [disabled]="processingId === u.id"
              >
                Enable
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  `,
  styles: [
    `
      .admin {
        max-width: 900px;
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
        grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
        gap: 0.75rem;
      }
      .card {
        background: rgba(15, 23, 42, 0.75);
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 12px;
        padding: 1rem;
      }
      .label {
        font-size: 0.75rem;
        color: #94a3b8;
        display: block;
      }
      .bad {
        color: #f87171;
      }
      h2 {
        font-size: 1.1rem;
        margin: 0.5rem 0 0;
      }
      .tbl {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.9rem;
      }
      .tbl th,
      .tbl td {
        border: 1px solid rgba(148, 163, 184, 0.2);
        padding: 0.5rem 0.65rem;
        text-align: left;
      }
      .tbl button {
        padding: 0.35rem 0.6rem;
        border-radius: 6px;
        border: 1px solid #22d3ee;
        background: #0f172a;
        color: #22d3ee;
        cursor: pointer;
      }
      .tbl button:disabled {
        opacity: 0.6;
        cursor: not-allowed;
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
export class AdminDashboardComponent implements OnInit {
  dash: AdminDashboard | null = null;
  dashError = '';
  users: AdminUserRow[] = [];
  usersLoading = false;
  usersError = '';
  processingId = '';

  constructor(private dashboardService: DashboardService) {}

  ngOnInit(): void {
    this.dashboardService.admin().subscribe({
      next: (d) => (this.dash = d),
      error: (err) => (this.dashError = err?.error?.error ?? 'Unable to load admin dashboard'),
    });
    this.loadUsers();
  }

  loadUsers(): void {
    this.usersLoading = true;
    this.usersError = '';
    this.dashboardService
      .listUsers()
      .pipe(finalize(() => (this.usersLoading = false)))
      .subscribe({
        next: (r) => (this.users = r.users ?? []),
        error: (err) => (this.usersError = err?.error?.error ?? 'Unable to load users'),
      });
  }

  toggleActive(u: AdminUserRow, active: boolean): void {
    this.processingId = u.id;
    this.dashboardService
      .setUserActive(u.id, active)
      .pipe(finalize(() => (this.processingId = '')))
      .subscribe({
        next: () => this.loadUsers(),
        error: (err) => (this.usersError = err?.error?.error ?? 'Update failed'),
      });
  }
}

import { Component, OnInit } from '@angular/core';
import { CommonModule, TitleCasePipe } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { AuthService, UserProfile } from '../../services/auth.service';

/** Account summary for patients and doctors (Sprint 3 feature 01 — Abhinav). */
@Component({
  selector: 'app-account-settings',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe],
  template: `
    <section class="account" data-cy="account-settings-page">
      <header class="toolbar">
        <button type="button" class="back" (click)="back()">{{ backLabel }}</button>
        <h1>Account</h1>
        <p class="lede">Your profile from the server. Use refresh if you updated your account elsewhere.</p>
      </header>

      <p *ngIf="loading" class="muted" data-cy="account-settings-loading">Loading profile…</p>
      <p *ngIf="error" class="error" data-cy="account-settings-error">{{ error }}</p>

      <div *ngIf="!loading && profile" class="card" data-cy="account-settings-card">
        <dl class="grid">
          <dt>Email</dt>
          <dd data-cy="account-settings-email">{{ profile.email }}</dd>
          <dt>Role</dt>
          <dd data-cy="account-settings-role">{{ profile.role | titlecase }}</dd>
          <dt>User ID</dt>
          <dd class="mono" data-cy="account-settings-id">{{ profile.id }}</dd>
        </dl>
        <button type="button" class="btn-secondary" (click)="reload()" [disabled]="loading" data-cy="account-settings-refresh">
          {{ loading ? 'Refreshing…' : 'Refresh from server' }}
        </button>
      </div>
    </section>
  `,
  styles: [`
    .account { max-width: 640px; margin: 0 auto; display: grid; gap: 1rem; padding: 0 0 1.5rem; }
    .toolbar { display: grid; gap: 0.35rem; }
    .back { width: fit-content; padding: 0.45rem 0.75rem; border-radius: 8px; border: 1px solid #64748b; background: #0f172a; color: #e2e8f0; cursor: pointer; }
    h1 { margin: 0; font-size: 1.35rem; color: #f1f5f9; }
    .lede { margin: 0; color: #94a3b8; font-size: 0.9rem; line-height: 1.45; }
    .card { background: rgba(15, 23, 42, 0.75); border: 1px solid rgba(148, 163, 184, 0.22); border-radius: 12px; padding: 1rem 1.1rem; display: grid; gap: 1rem; }
    .grid { display: grid; grid-template-columns: 7rem 1fr; gap: 0.5rem 0.75rem; margin: 0; font-size: 0.9rem; }
    dt { color: #64748b; margin: 0; }
    dd { margin: 0; color: #e2e8f0; }
    .mono { font-family: ui-monospace, monospace; font-size: 0.78rem; word-break: break-all; }
    .btn-secondary { width: fit-content; padding: 0.45rem 0.85rem; border-radius: 8px; border: 1px solid rgba(34, 211, 238, 0.35); background: rgba(15, 23, 42, 0.6); color: #a5f3fc; cursor: pointer; font-size: 0.85rem; }
    .btn-secondary:disabled { opacity: 0.55; cursor: not-allowed; }
    .error { color: #f87171; margin: 0; }
    .muted { color: #94a3b8; margin: 0; }
  `]
})
export class AccountSettingsComponent implements OnInit {
  profile: UserProfile | null = null;
  loading = false;
  error = '';

  constructor(
    private auth: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.reload();
  }

  get backLabel(): string {
    return this.auth.getUser()?.role === 'doctor' ? 'Back to queue' : 'Back to appointments';
  }

  back(): void {
    const role = this.auth.getUser()?.role;
    void this.router.navigate([role === 'doctor' ? '/doctor/appointments' : '/patient/appointments']);
  }

  reload(): void {
    this.loading = true;
    this.error = '';
    this.auth
      .refreshProfile()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (u) => (this.profile = u),
        error: (err) => {
          this.profile = this.auth.getUser();
          this.error = err?.error?.error ?? 'Unable to load profile from server.';
        }
      });
  }
}

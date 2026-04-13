import { Component, OnInit } from '@angular/core';
import { CommonModule, TitleCasePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { AuthService, UserProfile } from '../../services/auth.service';

@Component({
  selector: 'app-account-settings',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe],
  template: `
    <section class="account" data-cy="account-settings-page">
      <h1>Account</h1>
      <p class="subtitle">Profile data loaded from server via <code>/api/me</code>.</p>
      <div class="card">
        <p><strong>Email:</strong> <span data-cy="account-settings-email">{{ profile?.email || '—' }}</span></p>
        <p><strong>Role:</strong> <span data-cy="account-settings-role">{{ profile?.role | titlecase }}</span></p>
        <p><strong>User ID:</strong> <span>{{ profile?.id || '—' }}</span></p>
      </div>
      <button type="button" (click)="refresh()" [disabled]="loading">
        {{ loading ? 'Refreshing...' : 'Refresh from server' }}
      </button>
      <p *ngIf="error" class="error">{{ error }}</p>
    </section>
  `,
  styles: [`
    .account { max-width: 760px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle { color: #94a3b8; }
    .card { background: rgba(15,23,42,.7); border: 1px solid rgba(148,163,184,.2); border-radius: 12px; padding: 1rem; }
    .error { color: #f87171; }
    button { width: fit-content; padding: .55rem .85rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; cursor: pointer; }
  `]
})
export class AccountSettingsComponent implements OnInit {
  profile: UserProfile | null = null;
  loading = false;
  error = '';

  constructor(private auth: AuthService) {}

  ngOnInit(): void {
    this.profile = this.auth.getUser();
    this.refresh();
  }

  refresh(): void {
    this.loading = true;
    this.error = '';
    this.auth.refreshProfile().pipe(finalize(() => (this.loading = false))).subscribe({
      next: (profile) => (this.profile = profile),
      error: (err) => (this.error = err?.error?.error ?? 'Unable to refresh profile')
    });
  }
}


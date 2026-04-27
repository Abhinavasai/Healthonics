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

      <div class="summary" *ngIf="rows.length" data-cy="patient-prescriptions-summary">
        <div class="card">
          <span class="k">Active</span>
          <strong>{{ activeRows.length }}</strong>
        </div>
        <div class="card">
          <span class="k">History</span>
          <strong>{{ pastRows.length }}</strong>
        </div>
        <div class="card">
          <span class="k">Total</span>
          <strong>{{ rows.length }}</strong>
        </div>
      </div>

      <h2 *ngIf="activeRows.length">Active prescriptions</h2>
      <ul *ngIf="activeRows.length" data-cy="patient-prescriptions-list">
        <li *ngFor="let r of activeRows" data-cy="patient-prescriptions-row">
          <div class="head">
            <strong>{{ r.medication_name }}</strong>
            <span class="st st-active">{{ formatStatus(r.status) }}</span>
          </div>
          <div class="meta">{{ r.dosage }} · {{ r.frequency }} · {{ r.duration_days }} days</div>
          <div class="ins" *ngIf="r.instructions">{{ r.instructions }}</div>
          <div class="toggle-row">
            <small class="rem-status">
              Reminders: {{ remindersEnabled(r.id) ? 'On' : 'Off' }}
            </small>
            <button
              type="button"
              class="toggle-btn"
              (click)="toggleReminders(r.id)"
              data-cy="patient-prescriptions-reminder-toggle"
            >
              {{ remindersEnabled(r.id) ? 'Turn off reminders' : 'Turn on reminders' }}
            </button>
          </div>
          <small>Prescribed {{ r.created_at }}</small>
        </li>
      </ul>

      <h2 *ngIf="pastRows.length" class="mt">Prescription history</h2>
      <ul *ngIf="pastRows.length" data-cy="patient-prescriptions-history-list">
        <li *ngFor="let r of pastRows" data-cy="patient-prescriptions-history-row">
          <div class="head">
            <strong>{{ r.medication_name }}</strong>
            <span class="st st-past">{{ formatStatus(r.status) }}</span>
          </div>
          <div class="meta">{{ r.dosage }} · {{ r.frequency }} · {{ r.duration_days }} days</div>
          <div class="ins" *ngIf="r.instructions">{{ r.instructions }}</div>
          <small>Prescribed {{ r.created_at }}</small>
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
      .summary {
        display: grid;
        grid-template-columns: repeat(3, minmax(100px, 1fr));
        gap: 0.75rem;
        margin: 0.75rem 0 1rem;
      }
      .card {
        border: 1px solid rgba(148, 163, 184, 0.35);
        border-radius: 0.5rem;
        padding: 0.65rem 0.75rem;
      }
      .k {
        display: block;
        color: #64748b;
        font-size: 0.8rem;
      }
      ul {
        list-style: none;
        padding: 0;
        margin: 0;
        display: grid;
        gap: 0.6rem;
      }
      li {
        border: 1px solid rgba(148, 163, 184, 0.25);
        border-radius: 0.5rem;
        padding: 0.65rem 0.75rem;
      }
      .head {
        display: flex;
        justify-content: space-between;
        gap: 0.75rem;
        align-items: center;
      }
      .meta {
        color: #64748b;
        margin-top: 0.2rem;
      }
      .ins {
        margin-top: 0.25rem;
      }
      .toggle-row {
        margin-top: 0.35rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.75rem;
      }
      .rem-status {
        color: #64748b;
      }
      .toggle-btn {
        border: 1px solid rgba(148, 163, 184, 0.35);
        border-radius: 9999px;
        background: transparent;
        color: #0f172a;
        padding: 0.2rem 0.6rem;
        font-size: 0.78rem;
        cursor: pointer;
      }
      .toggle-btn:hover {
        background: rgba(148, 163, 184, 0.1);
      }
      .st {
        text-transform: uppercase;
        letter-spacing: 0.02em;
        font-size: 0.72rem;
        border-radius: 9999px;
        padding: 0.1rem 0.5rem;
      }
      .st-active {
        background: rgba(34, 197, 94, 0.15);
        color: #16a34a;
      }
      .st-past {
        background: rgba(148, 163, 184, 0.18);
        color: #475569;
      }
      .mt {
        margin-top: 1rem;
      }
    `
  ]
})
export class PatientPrescriptionsComponent implements OnInit {
  private readonly remindersStorageKey = 'prescriptionReminderPrefs';
  rows: PrescriptionRow[] = [];
  activeRows: PrescriptionRow[] = [];
  pastRows: PrescriptionRow[] = [];
  error: string | null = null;
  reminderToggles: Record<string, boolean> = {};

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
      next: (res) => {
        this.reminderToggles = this.loadReminderToggles();
        this.rows = [...(res.prescriptions ?? [])].sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
        this.activeRows = this.rows.filter((r) => r.status === 'active');
        this.pastRows = this.rows.filter((r) => r.status !== 'active');
      },
      error: () => (this.error = 'Could not load prescriptions')
    });
  }

  formatStatus(status: string): string {
    return (status || 'unknown').replace(/_/g, ' ');
  }

  remindersEnabled(id: string): boolean {
    return this.reminderToggles[id] ?? true;
  }

  toggleReminders(id: string): void {
    this.reminderToggles = {
      ...this.reminderToggles,
      [id]: !this.remindersEnabled(id)
    };
    this.saveReminderToggles();
  }

  private loadReminderToggles(): Record<string, boolean> {
    try {
      const raw = localStorage.getItem(this.remindersStorageKey);
      if (!raw) return {};
      const parsed = JSON.parse(raw) as unknown;
      if (!parsed || typeof parsed !== 'object') return {};
      return parsed as Record<string, boolean>;
    } catch {
      return {};
    }
  }

  private saveReminderToggles(): void {
    localStorage.setItem(this.remindersStorageKey, JSON.stringify(this.reminderToggles));
  }
}

import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { DashboardService, PatientDashboardSummary } from '../../services/dashboard.service';
import { AuthService } from '../../services/auth.service';
import { ApiContract } from '../../services/api-contract';

export interface AdherenceResult {
  score: number | null;
  label: string;
  total: number;
  active: number;
  revoked: number;
  completed: number;
  avg_days_used: number;
}

export interface BarRow {
  label: string;
  value: number;
  pct: number;
}

@Component({
  selector: 'app-patient-dashboard',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './patient-dashboard.component.html',
  // HttpClient provided globally via provideHttpClient() in app.config.ts
  styleUrl: './patient-dashboard.component.scss'
})
export class PatientDashboardComponent implements OnInit, OnDestroy {
  data: PatientDashboardSummary | null = null;
  error: string | null = null;
  adherence: AdherenceResult | null = null;
  fhirDownloading = false;
  fhirError: string | null = null;
  healthSummaryDownloading = false;

  /** Conic-gradient for the activity ring (non-zero total only). */
  pieBackground = 'conic-gradient(#334155 0% 100%)';

  private sub?: Subscription;
  private adherenceSub?: Subscription;

  constructor(
    private dash: DashboardService,
    private http: HttpClient,
    private auth: AuthService
  ) {}

  ngOnInit(): void {
    this.sub = this.dash.patientSummary().subscribe({
      next: (d) => {
        this.data = d;
        this.pieBackground = this.computePieGradient(d);
      },
      error: () => (this.error = 'Could not load summary')
    });
    const patientId = this.auth.getUser()?.id;
    if (patientId) {
      this.adherenceSub = this.http.get<AdherenceResult>(ApiContract.patients.adherenceScore(patientId)).subscribe({
        next: (res) => { this.adherence = res; },
        error: () => {}
      });
    }
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
    this.adherenceSub?.unsubscribe();
  }

  adherenceColor(label: string): string {
    const map: Record<string, string> = { Good: '#4ade80', Fair: '#facc15', Poor: '#f97316', 'Very Poor': '#f87171', 'No data': '#94a3b8' };
    return map[label] ?? '#94a3b8';
  }

  get barRows(): BarRow[] {
    const d = this.data;
    if (!d) {
      return [];
    }
    const agg = d.aggregations;
    const rows = [
      { label: 'Upcoming / active', value: d.upcoming_or_active_count },
      { label: 'Pending requests', value: d.pending_requests },
      { label: 'Unread messages', value: agg?.unread_messages ?? 0 },
      { label: 'Active prescriptions', value: agg?.active_prescriptions ?? 0 },
      { label: 'Documents', value: agg?.documents_uploaded ?? 0 }
    ];
    const max = Math.max(...rows.map((r) => r.value), 1);
    return rows.map((r) => ({
      ...r,
      pct: Math.round((r.value / max) * 100)
    }));
  }

  private computePieGradient(d: PatientDashboardSummary): string {
    const agg = d.aggregations;
    const parts: { color: string; value: number }[] = [
      { color: '#22c55e', value: d.upcoming_or_active_count },
      { color: '#eab308', value: d.pending_requests },
      { color: '#38bdf8', value: agg?.unread_messages ?? 0 },
      { color: '#a78bfa', value: (agg?.active_prescriptions ?? 0) + (agg?.documents_uploaded ?? 0) }
    ];
    const sum = parts.reduce((s, p) => s + p.value, 0);
    if (sum <= 0) {
      return 'conic-gradient(#334155 0% 100%)';
    }
    let acc = 0;
    const stops: string[] = [];
    for (const p of parts) {
      const pct = (p.value / sum) * 100;
      if (pct <= 0) {
        continue;
      }
      const start = acc;
      acc += pct;
      stops.push(`${p.color} ${start}% ${acc}%`);
    }
    return `conic-gradient(${stops.join(', ')})`;
  }

  alertIcon(sev: string): string {
    return sev === 'warning' ? '⚠' : 'ⓘ';
  }

  downloadHealthSummary(): void {
    if (this.healthSummaryDownloading) return;
    this.healthSummaryDownloading = true;
    this.http.get(ApiContract.patients.healthSummary, { responseType: 'blob' }).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `healthonyx-summary-${new Date().toISOString().slice(0, 10)}.txt`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
        this.healthSummaryDownloading = false;
      },
      error: () => { this.healthSummaryDownloading = false; }
    });
  }

  downloadFhir(): void {
    if (this.fhirDownloading) return;
    this.fhirDownloading = true;
    this.fhirError = null;
    this.http.get(ApiContract.patients.fhirExport, { responseType: 'blob' }).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `healthonyx-fhir-${new Date().toISOString().slice(0, 10)}.json`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
        this.fhirDownloading = false;
      },
      error: () => {
        this.fhirError = 'Could not download health record. Please try again.';
        this.fhirDownloading = false;
      }
    });
  }
}

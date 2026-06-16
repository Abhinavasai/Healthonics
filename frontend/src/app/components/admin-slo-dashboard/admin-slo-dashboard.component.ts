import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Subscription } from 'rxjs';
import { ApiContract } from '../../services/api-contract';

interface SLOData {
  latency_p50_ms: number | null;
  latency_p95_ms: number | null;
  latency_p99_ms: number | null;
  total_requests_24h: number;
  error_requests_24h: number;
  error_rate_pct: number;
  active_users_24h: number;
  recent_errors: { method: string; path: string; status_code: number; latency_ms: number; created_at: string }[];
  hourly_requests: { hour: string; requests: number }[];
}

@Component({
  selector: 'app-admin-slo-dashboard',
  standalone: true,
  imports: [CommonModule, DatePipe],
  template: `
    <section class="slo">
      <div class="slo-header">
        <h1>SLO Dashboard — Last 24h</h1>
        <button type="button" class="refresh-btn" (click)="load()" [disabled]="loading">Refresh</button>
      </div>
      <p *ngIf="loading && !data" class="muted">Loading...</p>
      <p *ngIf="error" class="error">{{ error }}</p>

      <div *ngIf="data" class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">P50 Latency</div>
          <div class="metric-value">{{ data.latency_p50_ms !== null ? (data.latency_p50_ms | number:'1.0-0') + ' ms' : '—' }}</div>
        </div>
        <div class="metric-card" [class.warn]="(data.latency_p95_ms ?? 0) > 500">
          <div class="metric-label">P95 Latency</div>
          <div class="metric-value">{{ data.latency_p95_ms !== null ? (data.latency_p95_ms | number:'1.0-0') + ' ms' : '—' }}</div>
        </div>
        <div class="metric-card" [class.danger]="(data.latency_p99_ms ?? 0) > 2000">
          <div class="metric-label">P99 Latency</div>
          <div class="metric-value">{{ data.latency_p99_ms !== null ? (data.latency_p99_ms | number:'1.0-0') + ' ms' : '—' }}</div>
        </div>
        <div class="metric-card" [class.danger]="data.error_rate_pct > 5">
          <div class="metric-label">Error Rate</div>
          <div class="metric-value">{{ data.error_rate_pct | number:'1.2-2' }}%</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Total Requests</div>
          <div class="metric-value">{{ data.total_requests_24h | number }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Active Users</div>
          <div class="metric-value">{{ data.active_users_24h }}</div>
        </div>
      </div>

      <div *ngIf="data && data.hourly_requests.length" class="section">
        <h2>Requests per Hour</h2>
        <div class="bar-chart">
          <div *ngFor="let h of data.hourly_requests" class="bar-col">
            <div class="bar" [style.height.px]="barHeight(h.requests)" [title]="h.hour + ': ' + h.requests"></div>
            <div class="bar-label">{{ h.hour | date:'HH' }}</div>
          </div>
        </div>
      </div>

      <div *ngIf="data && data.recent_errors.length" class="section">
        <h2>Recent 4xx/5xx Errors</h2>
        <table class="err-table">
          <thead>
            <tr><th>Method</th><th>Path</th><th>Status</th><th>Latency</th><th>Time</th></tr>
          </thead>
          <tbody>
            <tr *ngFor="let e of data.recent_errors" [class]="e.status_code >= 500 ? 'row-5xx' : 'row-4xx'">
              <td>{{ e.method }}</td>
              <td class="path">{{ e.path }}</td>
              <td class="status">{{ e.status_code }}</td>
              <td>{{ e.latency_ms }}ms</td>
              <td>{{ e.created_at | date:'HH:mm:ss' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  `,
  styles: [`
    .slo { max-width: 1100px; margin: 0 auto; padding: 1rem; }
    .slo-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.25rem; }
    h1 { margin: 0; font-size: 1.4rem; }
    h2 { font-size: 1rem; margin: 0 0 0.75rem; color: #cbd5e1; }
    .refresh-btn { padding: 0.4rem 0.9rem; border-radius: 8px; border: 1px solid #475569; background: #1e293b; color: #e2e8f0; cursor: pointer; }
    .muted { color: #94a3b8; }
    .error { color: #f87171; }
    .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 0.75rem; margin-bottom: 1.5rem; }
    .metric-card { background: rgba(15,23,42,0.8); border: 1px solid rgba(148,163,184,0.15); border-radius: 10px; padding: 0.85rem 1rem; }
    .metric-card.warn { border-color: rgba(251,191,36,0.4); }
    .metric-card.danger { border-color: rgba(248,113,113,0.4); }
    .metric-label { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.06em; color: #64748b; margin-bottom: 0.35rem; }
    .metric-value { font-size: 1.5rem; font-weight: 700; color: #e2e8f0; }
    .section { background: rgba(15,23,42,0.6); border: 1px solid rgba(148,163,184,0.1); border-radius: 10px; padding: 1rem; margin-bottom: 1rem; }
    .bar-chart { display: flex; align-items: flex-end; gap: 3px; height: 80px; }
    .bar-col { display: flex; flex-direction: column; align-items: center; gap: 3px; }
    .bar { width: 20px; background: #3b82f6; border-radius: 3px 3px 0 0; min-height: 2px; }
    .bar-label { font-size: 0.62rem; color: #64748b; }
    .err-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
    .err-table th { text-align: left; padding: 0.4rem 0.6rem; font-size: 0.75rem; text-transform: uppercase; color: #64748b; border-bottom: 1px solid #1e293b; }
    .err-table td { padding: 0.5rem 0.6rem; border-bottom: 1px solid rgba(30,41,59,0.5); color: #e2e8f0; }
    .path { font-family: monospace; color: #94a3b8; font-size: 0.82rem; }
    .status { font-weight: 700; }
    .row-5xx .status { color: #f87171; }
    .row-4xx .status { color: #fbbf24; }
  `]
})
export class AdminSloDashboardComponent implements OnInit, OnDestroy {
  data: SLOData | null = null;
  loading = false;
  error = '';
  private sub?: Subscription;

  constructor(private http: HttpClient) {}

  ngOnInit(): void { this.load(); }
  ngOnDestroy(): void { this.sub?.unsubscribe(); }

  load(): void {
    this.loading = true;
    this.sub?.unsubscribe();
    this.sub = this.http.get<SLOData>(ApiContract.admin.slo).subscribe({
      next: (d) => { this.data = d; this.loading = false; },
      error: () => { this.error = 'Could not load SLO data'; this.loading = false; }
    });
  }

  barHeight(requests: number): number {
    const max = Math.max(...(this.data?.hourly_requests.map(h => h.requests) ?? [1]));
    return max > 0 ? Math.max(2, Math.round((requests / max) * 70)) : 2;
  }
}

import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { DashboardService, DoctorCriticalEscalation, DoctorDashboardSummary } from '../../services/dashboard.service';
import { DoctorWorkloadHeatmapComponent } from '../doctor-workload-heatmap/doctor-workload-heatmap.component';

export interface DoctorBarRow {
  label: string;
  value: number;
  pct: number;
}

@Component({
  selector: 'app-doctor-dashboard',
  standalone: true,
  imports: [CommonModule, DoctorWorkloadHeatmapComponent],
  templateUrl: './doctor-dashboard.component.html',
  styleUrl: './doctor-dashboard.component.scss'
})
export class DoctorDashboardComponent implements OnInit, OnDestroy {
  data: DoctorDashboardSummary | null = null;
  error: string | null = null;
  ackingEscalationId: string | null = null;
  pieBackground = 'conic-gradient(#334155 0% 100%)';

  private sub?: Subscription;

  constructor(private dash: DashboardService) {}

  ngOnInit(): void {
    this.sub = this.dash.doctorSummary().subscribe({
      next: (d) => {
        this.data = d;
        this.pieBackground = this.computePieGradient(d);
      },
      error: () => (this.error = 'Could not load summary')
    });
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  get barRows(): DoctorBarRow[] {
    const d = this.data;
    if (!d) {
      return [];
    }
    const agg = d.aggregations;
    const rows = [
      { label: 'Appointments today', value: d.appointments_today },
      { label: 'Pending queue', value: d.pending_queue },
      { label: 'Unread messages', value: agg?.unread_messages ?? 0 },
      { label: 'Scheduled this week', value: agg?.appointments_this_week ?? 0 }
    ];
    const max = Math.max(...rows.map((r) => r.value), 1);
    return rows.map((r) => ({
      ...r,
      pct: Math.round((r.value / max) * 100)
    }));
  }

  private computePieGradient(d: DoctorDashboardSummary): string {
    const agg = d.aggregations;
    const parts: { color: string; value: number }[] = [
      { color: '#f59e0b', value: d.appointments_today },
      { color: '#f43f5e', value: d.pending_queue },
      { color: '#38bdf8', value: agg?.unread_messages ?? 0 },
      { color: '#a78bfa', value: agg?.appointments_this_week ?? 0 }
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

  acknowledgeEscalation(row: DoctorCriticalEscalation): void {
    if (this.ackingEscalationId) {
      return;
    }
    this.ackingEscalationId = row.id;
    this.dash.acknowledgeCriticalEscalation(row.id).subscribe({
      next: () => {
        if (!this.data?.critical_escalations) {
          this.ackingEscalationId = null;
          return;
        }
        this.data = {
          ...this.data,
          critical_escalations: this.data.critical_escalations.map((r) =>
            r.id === row.id ? { ...r, status: 'acknowledged', acknowledged_at: new Date().toISOString() } : r
          )
        };
        this.ackingEscalationId = null;
      },
      error: () => {
        this.ackingEscalationId = null;
        this.error = 'Could not acknowledge escalation. Please try again.';
      }
    });
  }
}

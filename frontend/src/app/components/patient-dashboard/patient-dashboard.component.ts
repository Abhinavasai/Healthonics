import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DashboardService, PatientDashboardSummary } from '../../services/dashboard.service';

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
  styleUrl: './patient-dashboard.component.scss'
})
export class PatientDashboardComponent implements OnInit {
  data: PatientDashboardSummary | null = null;
  error: string | null = null;

  /** Conic-gradient for the activity ring (non-zero total only). */
  pieBackground = 'conic-gradient(#334155 0% 100%)';

  constructor(private dash: DashboardService) {}

  ngOnInit(): void {
    this.dash.patientSummary().subscribe({
      next: (d) => {
        this.data = d;
        this.pieBackground = this.computePieGradient(d);
      },
      error: () => (this.error = 'Could not load summary')
    });
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
}

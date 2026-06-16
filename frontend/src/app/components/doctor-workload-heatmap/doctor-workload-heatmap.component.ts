import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Subscription } from 'rxjs';
import { ApiContract } from '../../services/api-contract';

interface DayCount { day: string; count: number; }
interface Cell { date: Date; count: number; label: string; }

@Component({
  selector: 'app-doctor-workload-heatmap',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="heatmap-wrap">
      <div class="heatmap-header">
        <h3>Appointment Workload — Last 90 Days</h3>
        <div class="legend">
          <span class="leg-label">Less</span>
          <span class="leg-cell l0"></span>
          <span class="leg-cell l1"></span>
          <span class="leg-cell l2"></span>
          <span class="leg-cell l3"></span>
          <span class="leg-cell l4"></span>
          <span class="leg-label">More</span>
        </div>
      </div>
      <p *ngIf="loading" class="muted">Loading heatmap...</p>
      <div *ngIf="!loading" class="heatmap-grid">
        <div *ngFor="let week of weeks" class="week-col">
          <div *ngFor="let cell of week" class="day-cell"
               [class]="'level-' + intensity(cell.count)"
               [title]="cell.label">
          </div>
        </div>
      </div>
      <div class="month-labels" *ngIf="!loading">
        <span *ngFor="let m of monthLabels" class="month-label" [style.margin-left.px]="m.offset">{{ m.label }}</span>
      </div>
      <p *ngIf="!loading && maxCount === 0" class="muted empty-msg">No appointments in the last 90 days.</p>
      <p *ngIf="!loading && maxCount > 0" class="peak-info">Peak day: {{ maxCount }} appointments</p>
    </div>
  `,
  styles: [`
    .heatmap-wrap { padding: 16px 0; }
    .heatmap-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; flex-wrap: wrap; gap: 8px; }
    h3 { margin: 0; font-size: 1rem; font-weight: 700; color: #e2e8f0; }
    .legend { display: flex; align-items: center; gap: 4px; }
    .leg-label { font-size: 0.72rem; color: #64748b; }
    .leg-cell { width: 12px; height: 12px; border-radius: 2px; }
    .l0 { background: #1e293b; }
    .l1 { background: #1e3a5f; }
    .l2 { background: #1d4ed8; }
    .l3 { background: #3b82f6; }
    .l4 { background: #93c5fd; }
    .heatmap-grid { display: flex; gap: 3px; overflow-x: auto; padding-bottom: 4px; }
    .week-col { display: flex; flex-direction: column; gap: 3px; }
    .day-cell { width: 12px; height: 12px; border-radius: 2px; cursor: default; transition: transform 0.1s; }
    .day-cell:hover { transform: scale(1.4); }
    .level-0 { background: #1e293b; }
    .level-1 { background: #1e3a5f; }
    .level-2 { background: #1d4ed8; }
    .level-3 { background: #3b82f6; }
    .level-4 { background: #93c5fd; }
    .month-labels { display: flex; margin-top: 4px; position: relative; }
    .month-label { font-size: 0.68rem; color: #64748b; position: absolute; white-space: nowrap; }
    .muted { color: #94a3b8; font-size: 0.85rem; }
    .empty-msg { margin-top: 8px; }
    .peak-info { font-size: 0.75rem; color: #64748b; margin-top: 6px; }
  `]
})
export class DoctorWorkloadHeatmapComponent implements OnInit, OnDestroy {
  loading = true;
  weeks: Cell[][] = [];
  maxCount = 0;
  monthLabels: { label: string; offset: number }[] = [];
  private sub?: Subscription;

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.sub = this.http.get<{ data: DayCount[]; max_count: number }>(ApiContract.doctor.workloadHeatmap).subscribe({
      next: (res) => {
        this.maxCount = res.max_count ?? 0;
        this.buildGrid(res.data ?? []);
        this.loading = false;
      },
      error: () => { this.loading = false; }
    });
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  intensity(count: number): number {
    if (count === 0 || this.maxCount === 0) return 0;
    const ratio = count / this.maxCount;
    if (ratio <= 0.25) return 1;
    if (ratio <= 0.5) return 2;
    if (ratio <= 0.75) return 3;
    return 4;
  }

  private buildGrid(data: DayCount[]): void {
    const countMap = new Map<string, number>();
    for (const d of data) countMap.set(d.day, d.count);

    const today = new Date();
    const start = new Date(today);
    start.setDate(start.getDate() - 89);
    // Go back to the start of the week (Sunday).
    start.setDate(start.getDate() - start.getDay());

    const weeks: Cell[][] = [];
    const monthSet = new Map<string, number>();
    let weekIndex = 0;

    const cur = new Date(start);
    while (cur <= today) {
      const week: Cell[] = [];
      for (let d = 0; d < 7; d++) {
        const iso = cur.toISOString().slice(0, 10);
        const count = countMap.get(iso) ?? 0;
        const monthKey = cur.toLocaleString('default', { month: 'short' });
        if (cur.getDate() <= 7 && !monthSet.has(monthKey)) {
          monthSet.set(monthKey, weekIndex * 15);
        }
        week.push({
          date: new Date(cur),
          count,
          label: `${iso}: ${count} appointment${count !== 1 ? 's' : ''}`
        });
        cur.setDate(cur.getDate() + 1);
      }
      weeks.push(week);
      weekIndex++;
    }
    this.weeks = weeks;
    this.monthLabels = Array.from(monthSet.entries()).map(([label, offset]) => ({ label, offset }));
  }
}

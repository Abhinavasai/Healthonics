import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Subscription, timer } from 'rxjs';
import { switchMap } from 'rxjs/operators';
import { ApiContract } from '../../services/api-contract';

interface QueueEntry {
  appointment_id: string;
  patient_email: string;
  scheduled_at: string;
  checked_in_at: string | null;
  reason: string;
  status: string;
  wait_minutes: number | null;
}

@Component({
  selector: 'app-waiting-room',
  standalone: true,
  imports: [CommonModule, DatePipe],
  template: `
    <section class="wr">
      <div class="wr-header">
        <h1>Waiting Room — Today</h1>
        <span class="live-badge">Live</span>
      </div>
      <p *ngIf="loading && !queue.length" class="muted">Loading...</p>
      <p *ngIf="error" class="error">{{ error }}</p>
      <div *ngIf="!queue.length && !loading" class="empty">No patients checked in or scheduled for today.</div>
      <table *ngIf="queue.length" class="wt">
        <thead>
          <tr>
            <th>#</th>
            <th>Patient</th>
            <th>Scheduled</th>
            <th>Check-in</th>
            <th>Wait</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let e of queue; let i = index" [class]="'row-' + e.status">
            <td class="pos">{{ i + 1 }}</td>
            <td>{{ e.patient_email }}</td>
            <td>{{ e.scheduled_at | date:'shortTime' }}</td>
            <td>{{ e.checked_in_at ? (e.checked_in_at | date:'shortTime') : '—' }}</td>
            <td class="wait">{{ e.wait_minutes !== null ? e.wait_minutes + ' min' : '—' }}</td>
            <td><span class="st-badge" [class]="'st-' + e.status">{{ e.status }}</span></td>
          </tr>
        </tbody>
      </table>
      <p class="refresh-note">Refreshes every 30 seconds</p>
    </section>
  `,
  styles: [`
    .wr { max-width: 900px; margin: 0 auto; padding: 1rem; }
    .wr-header { display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem; }
    h1 { margin: 0; font-size: 1.4rem; }
    .live-badge { padding: 2px 10px; border-radius: 999px; background: #22c55e; color: #fff; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; animation: pulse 2s infinite; }
    @keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.5; } }
    .muted { color: #94a3b8; }
    .error { color: #f87171; }
    .empty { color: #64748b; padding: 2rem 0; text-align: center; }
    .wt { width: 100%; border-collapse: collapse; }
    .wt th { text-align: left; padding: 0.6rem 0.75rem; font-size: 0.78rem; text-transform: uppercase; letter-spacing: 0.05em; color: #64748b; border-bottom: 1px solid #1e293b; }
    .wt td { padding: 0.65rem 0.75rem; border-bottom: 1px solid rgba(30,41,59,0.6); color: #e2e8f0; font-size: 0.9rem; }
    .pos { font-weight: 700; color: #94a3b8; text-align: center; }
    .wait { font-weight: 700; color: #fbbf24; }
    .st-badge { border-radius: 6px; padding: 2px 8px; font-size: 0.76rem; font-weight: 600; }
    .st-approved { background: rgba(34,197,94,0.15); color: #4ade80; }
    .st-completed { background: rgba(148,163,184,0.1); color: #94a3b8; }
    .refresh-note { font-size: 0.75rem; color: #475569; margin-top: 1rem; }
  `]
})
export class WaitingRoomComponent implements OnInit, OnDestroy {
  queue: QueueEntry[] = [];
  loading = false;
  error = '';
  private sub?: Subscription;

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.loading = true;
    this.sub = timer(0, 30000).pipe(
      switchMap(() => this.http.get<{ queue: QueueEntry[] }>(ApiContract.doctor.waitingRoom))
    ).subscribe({
      next: (res) => { this.queue = res.queue ?? []; this.loading = false; },
      error: () => { this.error = 'Could not load waiting room'; this.loading = false; }
    });
  }

  ngOnDestroy(): void { this.sub?.unsubscribe(); }
}

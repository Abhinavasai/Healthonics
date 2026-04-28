import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { Appointment, AppointmentComment, AppointmentsService } from '../../services/appointments.service';

@Component({
  selector: 'app-patient-appointment-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, DatePipe, TitleCasePipe],
  template: `
    <section class="detail">
      <header class="toolbar">
        <button type="button" class="back" (click)="back()">Back to list</button>
        <h1>Appointment details</h1>
      </header>
      <p *ngIf="loading" class="muted">Loading...</p>
      <p *ngIf="error" class="error">{{ error }}</p>
      <div *ngIf="!loading && appointment" class="card">
        <div class="top-row">
          <strong>{{ appointment.scheduled_at | date: 'medium' }}</strong>
          <span class="badge" [class]="'badge ' + appointment.status">
            {{ appointment.status | titlecase }}
          </span>
        </div>
        <div class="meta">Doctor: {{ appointment.doctor_id }}</div>
        <div class="reason">{{ appointment.reason }}</div>
      </div>

      <div *ngIf="!loading && appointment" class="card" data-cy="appointment-comments-section">
        <h2>Comments</h2>
        <ul *ngIf="comments.length" data-cy="appointment-comments-list">
          <li *ngFor="let c of comments">
            {{ c.body }} — <small>{{ c.created_at }}</small>
          </li>
        </ul>
        <p *ngIf="!comments.length" class="muted">No comments yet.</p>
        <label class="cmt">
          Add comment
          <textarea [(ngModel)]="newComment" rows="2" data-cy="appointment-comment-input"></textarea>
        </label>
        <button type="button" (click)="submitComment()" [disabled]="!newComment.trim() || posting" data-cy="appointment-comment-submit">
          Post
        </button>
      </div>
    </section>
  `,
  styles: [`
    .detail { max-width: 720px; margin: 0 auto; display: grid; gap: 1rem; }
    .toolbar { display: flex; flex-direction: column; gap: 0.35rem; }
    .back { width: fit-content; padding: 0.45rem 0.75rem; border-radius: 8px; border: 1px solid #64748b; background: #0f172a; color: #e2e8f0; cursor: pointer; }
    h1 { margin: 0; font-size: 1.35rem; }
    .card { background: rgba(15, 23, 42, 0.7); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 12px; padding: 1rem; }
    .top-row { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
    .badge { border-radius: 999px; padding: 0.15rem 0.55rem; font-size: 0.78rem; border: 1px solid transparent; }
    .pending { color: #facc15; border-color: rgba(250, 204, 21, 0.5); }
    .approved { color: #22c55e; border-color: rgba(34, 197, 94, 0.5); }
    .rejected { color: #f87171; border-color: rgba(248, 113, 113, 0.5); }
    .meta { color: #94a3b8; margin-top: 0.35rem; font-size: 0.9rem; }
    .reason { margin-top: 0.5rem; }
      .error { color: #f87171; }
    .muted { color: #94a3b8; }
    .cmt { display: grid; gap: 0.35rem; margin-top: 0.75rem; }
    ul { margin: 0.5rem 0; padding-left: 1.1rem; }
  `]
})
export class PatientAppointmentDetailComponent implements OnInit {
  appointment: Appointment | null = null;
  comments: AppointmentComment[] = [];
  newComment = '';
  posting = false;
  loading = false;
  error = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private appointmentsService: AppointmentsService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.error = 'Missing appointment id';
      return;
    }
    this.load(id);
  }

  load(id: string): void {
    this.loading = true;
    this.error = '';
    this.appointmentsService
      .getById(id)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (appt) => {
          this.appointment = appt;
          this.reloadComments(id);
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Unable to load appointment';
        }
      });
  }

  reloadComments(id: string): void {
    this.appointmentsService.listComments(id).subscribe({
      next: (r) => (this.comments = r.comments ?? []),
      error: () => {}
    });
  }

  submitComment(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id || !this.appointment) {
      return;
    }
    const body = this.newComment.trim();
    if (!body) {
      return;
    }
    this.posting = true;
    this.appointmentsService
      .postComment(id, body)
      .pipe(finalize(() => (this.posting = false)))
      .subscribe({
        next: () => {
          this.newComment = '';
          this.reloadComments(id);
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Could not post comment';
        }
      });
  }

  back(): void {
    void this.router.navigate(['/patient/appointments']);
  }
}

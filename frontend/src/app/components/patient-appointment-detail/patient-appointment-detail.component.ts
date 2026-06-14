import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { finalize } from 'rxjs';
import { Appointment, AppointmentComment, AppointmentsService } from '../../services/appointments.service';
import { ApiContract } from '../../services/api-contract';
import { AppointmentRatingComponent } from '../appointment-rating/appointment-rating.component';

interface QuestionnaireState {
  questions: string[];
  answers: string[];
  submitted_at: string | null;
}

@Component({
  selector: 'app-patient-appointment-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, DatePipe, TitleCasePipe, AppointmentRatingComponent],
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
        <div class="meta">Doctor: {{ appointment.doctor_email || appointment.doctor_id }}</div>
        <div class="reason">{{ appointment.reason }}</div>

        <!-- Video consult link -->
        <div *ngIf="appointment.video_link && appointment.status === 'approved'" class="video-link-banner">
          <span>Video Consultation:</span>
          <a [href]="appointment.video_link" target="_blank" rel="noopener" class="video-link-btn">
            Join Video Call
          </a>
        </div>

        <!-- Check-in button -->
        <div *ngIf="appointment.status === 'approved'" style="margin-top:0.75rem">
          <button *ngIf="!appointment.checked_in_at" type="button" class="btn-checkin"
            (click)="checkIn()" [disabled]="checkingIn">
            {{ checkingIn ? 'Checking in...' : 'Check In for Today\'s Appointment' }}
          </button>
          <span *ngIf="appointment.checked_in_at" class="checked-in-badge">✓ Checked in</span>
        </div>
      </div>

      <!-- Pre-Visit Questionnaire -->
      <div *ngIf="!loading && appointment" class="card q-card">
        <div class="q-header">
          <h2>Pre-Visit Intake Form</h2>
          <span *ngIf="questionnaire?.submitted_at" class="q-submitted">Submitted ✓</span>
        </div>
        <p *ngIf="qLoading" class="muted">Loading questions...</p>
        <div *ngIf="questionnaire && !qLoading">
          <div *ngIf="questionnaire.submitted_at" class="q-done">
            Your responses have been shared with your doctor. Thank you!
          </div>
          <div *ngIf="!questionnaire.submitted_at">
            <div *ngFor="let q of questionnaire.questions; let i = index" class="q-item">
              <label class="q-label">{{ i + 1 }}. {{ q }}</label>
              <textarea
                [(ngModel)]="questionnaire.answers[i]"
                rows="2"
                placeholder="Your answer..."
                [disabled]="submittingAnswers"
              ></textarea>
            </div>
            <button class="btn-submit-q" (click)="submitAnswers()" [disabled]="submittingAnswers || !answersComplete()">
              {{ submittingAnswers ? 'Submitting...' : 'Submit Answers' }}
            </button>
            <p *ngIf="qError" class="error">{{ qError }}</p>
          </div>
        </div>
      </div>

      <!-- Rating -->
      <app-appointment-rating
        *ngIf="!loading && appointment && appointment.status === 'completed'"
        [appointmentId]="appointment.id"
        [status]="appointment.status">
      </app-appointment-rating>

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
    .q-card h2 { margin: 0 0 0.75rem; }
    .q-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem; }
    .q-header h2 { margin: 0; }
    .q-submitted { font-size: 0.78rem; color: #4ade80; background: rgba(74, 222, 128, 0.1); border: 1px solid rgba(74,222,128,0.3); border-radius: 20px; padding: 2px 10px; }
    .q-done { color: #4ade80; font-size: 0.9rem; padding: 10px 0; }
    .q-item { margin-bottom: 0.75rem; }
    .q-label { display: block; font-size: 0.88rem; color: #cbd5e1; margin-bottom: 4px; font-weight: 500; }
    textarea { width: 100%; background: rgba(255,255,255,0.05); border: 1px solid rgba(148,163,184,0.25); border-radius: 8px; padding: 8px 10px; color: #e2e8f0; font-family: inherit; font-size: 0.88rem; resize: vertical; box-sizing: border-box; }
    textarea:focus { outline: none; border-color: #4f46e5; }
    .btn-submit-q { margin-top: 0.5rem; padding: 0.5rem 1.25rem; background: #4f46e5; color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; }
    .btn-submit-q:disabled { opacity: 0.5; cursor: not-allowed; }
    .video-link-banner { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.75rem; padding: 0.6rem 0.85rem; background: rgba(59,130,246,0.1); border: 1px solid rgba(59,130,246,0.3); border-radius: 8px; }
    .video-link-btn { padding: 0.35rem 0.85rem; background: #3b82f6; color: #fff; border-radius: 6px; text-decoration: none; font-weight: 600; font-size: 0.88rem; }
    .btn-checkin { padding: 0.45rem 1rem; background: #22c55e; color: #fff; border: none; border-radius: 8px; cursor: pointer; font-weight: 600; }
    .btn-checkin:disabled { opacity: 0.5; cursor: not-allowed; }
    .checked-in-badge { color: #4ade80; font-size: 0.88rem; font-weight: 600; }
  `]
})
export class PatientAppointmentDetailComponent implements OnInit {
  appointment: Appointment | null = null;
  comments: AppointmentComment[] = [];
  newComment = '';
  posting = false;
  loading = false;
  error = '';

  questionnaire: QuestionnaireState | null = null;
  qLoading = false;
  qError = '';
  submittingAnswers = false;

  checkingIn = false;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private appointmentsService: AppointmentsService,
    private http: HttpClient
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) { this.error = 'Missing appointment id'; return; }
    this.load(id);
  }

  load(id: string): void {
    this.loading = true;
    this.error = '';
    this.appointmentsService.getById(id).pipe(finalize(() => (this.loading = false))).subscribe({
      next: (appt) => {
        this.appointment = appt;
        this.reloadComments(id);
        this.loadQuestionnaire(id);
      },
      error: (err) => { this.error = err?.error?.error ?? 'Unable to load appointment'; }
    });
  }

  loadQuestionnaire(id: string): void {
    this.qLoading = true;
    this.http.get<{ questions: string; answers: string; submitted_at: string | null }>(
      `${ApiContract.appointments.byId(id)}/questionnaire`
    ).pipe(finalize(() => (this.qLoading = false))).subscribe({
      next: (res) => {
        let questions: string[] = [];
        let answers: string[] = [];
        try { questions = JSON.parse(res.questions); } catch {}
        try { if (res.answers) answers = JSON.parse(res.answers); } catch {}
        while (answers.length < questions.length) answers.push('');
        this.questionnaire = { questions, answers, submitted_at: res.submitted_at };
      },
      error: () => {}
    });
  }

  answersComplete(): boolean {
    return (this.questionnaire?.answers ?? []).every(a => a.trim().length > 0);
  }

  submitAnswers(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id || !this.questionnaire) return;
    this.submittingAnswers = true;
    this.qError = '';
    const answers = JSON.stringify(this.questionnaire.answers);
    this.http.post(`${ApiContract.appointments.byId(id)}/questionnaire`, { answers }).pipe(
      finalize(() => (this.submittingAnswers = false))
    ).subscribe({
      next: () => {
        if (this.questionnaire) this.questionnaire.submitted_at = new Date().toISOString();
      },
      error: (err) => { this.qError = err?.error?.error ?? 'Could not submit answers'; }
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
    if (!id || !this.appointment) return;
    const body = this.newComment.trim();
    if (!body) return;
    this.posting = true;
    this.appointmentsService.postComment(id, body).pipe(finalize(() => (this.posting = false))).subscribe({
      next: () => { this.newComment = ''; this.reloadComments(id); },
      error: (err) => { this.error = err?.error?.error ?? 'Could not post comment'; }
    });
  }

  checkIn(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) return;
    this.checkingIn = true;
    this.http.post(ApiContract.appointments.checkIn(id), {}).subscribe({
      next: () => {
        if (this.appointment) this.appointment = { ...this.appointment, checked_in_at: new Date().toISOString() };
        this.checkingIn = false;
      },
      error: () => { this.checkingIn = false; }
    });
  }

  back(): void { void this.router.navigate(['/patient/appointments']); }
}

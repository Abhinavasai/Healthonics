import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { finalize, Subject, debounceTime, distinctUntilChanged, switchMap, takeUntil } from 'rxjs';
import {
  Appointment,
  AppointmentComment,
  AppointmentsService,
  CommentVisibility
} from '../../services/appointments.service';
import { HttpClient } from '@angular/common/http';
import { AssistantService, PatientSummaryResponse, NoteAssistResponse } from '../../services/assistant.service';
import { ApiContract } from '../../services/api-contract';

@Component({
  selector: 'app-doctor-appointment-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, DatePipe, TitleCasePipe],
  template: `
    <section class="detail">
      <header class="toolbar">
        <button type="button" class="back" (click)="back()">Back to queue</button>
        <h1>Request details</h1>
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
        <div class="meta" style="display:flex;align-items:center;gap:0.75rem;flex-wrap:wrap">
          Patient: {{ appointment.patient_id }}
          <span *ngIf="adherence" class="adh-badge" [class]="'adh-' + adherence.label.toLowerCase().replace(' ','-')">
            Adherence: {{ adherence.score }}% · {{ adherence.label }}
          </span>
        </div>
        <div class="reason">{{ appointment.reason }}</div>
      </div>

      <!-- Patient Pre-Visit Questionnaire Answers -->
      <div *ngIf="!loading && patientAnswers.length" class="card">
        <h2 style="margin-top:0">Patient Intake Responses</h2>
        <div *ngFor="let qa of patientAnswers" class="qa-row">
          <div class="qa-q">{{ qa.question }}</div>
          <div class="qa-a">{{ qa.answer }}</div>
        </div>
      </div>

      <!-- Teleconsult Video Link -->
      <div *ngIf="!loading && appointment" class="card">
        <h2 style="margin-top:0">Video Consultation Link</h2>
        <p class="muted" style="font-size:0.82rem;margin-bottom:0.75rem">
          Set a Jitsi Meet or other video call URL. The patient will see this link before their visit.
        </p>
        <div style="display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap">
          <input type="url" [(ngModel)]="videoLinkInput" placeholder="https://meet.jit.si/your-room"
            style="flex:1;min-width:200px;background:#0f172a;border:1px solid #475569;color:#e2e8f0;border-radius:8px;padding:0.45rem 0.75rem;font-size:0.9rem" />
          <button type="button" (click)="saveVideoLink()" [disabled]="videoLinkSaving">
            {{ videoLinkSaving ? 'Saving…' : 'Save Link' }}
          </button>
          <button type="button" *ngIf="appointment.video_link" (click)="clearVideoLink()" [disabled]="videoLinkSaving"
            style="background:#1e293b;border-color:#ef4444;color:#ef4444">Clear</button>
        </div>
        <p *ngIf="videoLinkError" class="error" style="margin-top:0.5rem">{{ videoLinkError }}</p>
        <p *ngIf="appointment.video_link" style="margin-top:0.5rem;font-size:0.85rem">
          Current: <a [href]="appointment.video_link" target="_blank" rel="noopener" style="color:#60a5fa">{{ appointment.video_link }}</a>
        </p>
      </div>

      <!-- AI Patient Summary -->
      <div *ngIf="!loading && appointment" class="card ai-summary-card">
        <div class="ai-summary-header">
          <h2>AI Pre-Visit Briefing</h2>
          <button type="button" class="btn-ai" (click)="loadAiSummary()" [disabled]="aiLoading">
            <span *ngIf="!aiLoading">{{ aiSummary ? 'Refresh' : 'Generate Summary' }}</span>
            <span *ngIf="aiLoading">Generating...</span>
          </button>
        </div>
        <div *ngIf="aiError" class="error">{{ aiError }}</div>
        <div *ngIf="aiSummary" class="ai-summary-body">
          <pre>{{ aiSummary.summary }}</pre>
          <div class="ai-meta">
            Provider: {{ aiSummary.provider }}{{ aiSummary.fallback_used ? ' (fallback)' : '' }}
          </div>
        </div>
        <p *ngIf="!aiSummary && !aiLoading && !aiError" class="muted">
          Generate an AI briefing of this patient's recent history before the visit.
        </p>
      </div>

      <div *ngIf="!loading && appointment" class="card" data-cy="doctor-appointment-comments-section">
        <h2>Comments</h2>
        <ul *ngIf="comments.length" class="comment-list" data-cy="doctor-appointment-comments-list">
          <li *ngFor="let c of comments">
            <span *ngIf="c.visibility === 'internal'" class="badge-internal" data-cy="comment-internal-badge"
              >Internal</span
            >
            {{ c.body }} — <small>{{ c.created_at }}</small>
          </li>
        </ul>
        <p *ngIf="!comments.length" class="muted">No comments yet.</p>
        <div class="visibility-pick" data-cy="comment-visibility">
          <span class="vis-label">Visibility</span>
          <label class="vis-opt">
            <input type="radio" name="commentVis" [(ngModel)]="commentVisibility" value="patient_visible" />
            Visible to patient
          </label>
          <label class="vis-opt">
            <input type="radio" name="commentVis" [(ngModel)]="commentVisibility" value="internal" />
            Internal (care team only)
          </label>
        </div>
        <label class="cmt">
          Add comment
          <textarea [(ngModel)]="newComment" (ngModelChange)="onNoteInput()" rows="4" data-cy="doctor-appointment-comment-input" placeholder="Start typing your clinical note... AI will suggest ICD-10 codes and a SOAP scaffold as you write."></textarea>
        </label>

        <!-- AI Note Assist Sidebar -->
        <div *ngIf="noteAssistLoading" class="note-assist-loading">AI is analysing your note...</div>
        <div *ngIf="noteAssist && !noteAssistLoading" class="note-assist-panel">
          <div class="na-header">
            <span class="na-badge">AI Note Assistant</span>
            <span class="na-provider">{{ noteAssist.provider }}</span>
          </div>
          <div class="na-section" *ngIf="noteAssist.icd10_suggestions">
            <div class="na-label">ICD-10 Suggestions</div>
            <div class="na-value">{{ noteAssist.icd10_suggestions }}</div>
          </div>
          <div class="na-section" *ngIf="noteAssist.soap_scaffold">
            <div class="na-label">SOAP Scaffold</div>
            <pre class="na-soap">{{ noteAssist.soap_scaffold }}</pre>
          </div>
          <div class="na-section" *ngIf="noteAssist.clinical_flags">
            <div class="na-label">Clinical Flags</div>
            <div class="na-value na-flags">{{ noteAssist.clinical_flags }}</div>
          </div>
        </div>
        <button
          type="button"
          (click)="submitComment()"
          [disabled]="!newComment.trim() || posting"
          data-cy="doctor-appointment-comment-submit"
        >
          Post
        </button>
      </div>
    </section>
  `,
  styles: [
    `
      .detail {
        max-width: 720px;
        margin: 0 auto;
        display: grid;
        gap: 1rem;
      }
      .toolbar {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
      }
      .back {
        width: fit-content;
        padding: 0.45rem 0.75rem;
        border-radius: 8px;
        border: 1px solid #64748b;
        background: #0f172a;
        color: #e2e8f0;
        cursor: pointer;
      }
      h1 {
        margin: 0;
        font-size: 1.35rem;
      }
      .card {
        background: rgba(15, 23, 42, 0.7);
        border: 1px solid rgba(148, 163, 184, 0.2);
        border-radius: 12px;
        padding: 1rem;
      }
      .top-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 0.5rem;
      }
      .badge {
        border-radius: 999px;
        padding: 0.15rem 0.55rem;
        font-size: 0.78rem;
        border: 1px solid transparent;
      }
      .pending {
        color: #facc15;
        border-color: rgba(250, 204, 21, 0.5);
      }
      .approved {
        color: #22c55e;
        border-color: rgba(34, 197, 94, 0.5);
      }
      .rejected {
        color: #f87171;
        border-color: rgba(248, 113, 113, 0.5);
      }
      .meta {
        color: #94a3b8;
        margin-top: 0.35rem;
        font-size: 0.9rem;
      }
      .reason {
        margin-top: 0.5rem;
      }
      .error {
        color: #f87171;
      }
      .muted {
        color: #94a3b8;
      }
      .cmt {
        display: grid;
        gap: 0.35rem;
        margin-top: 0.75rem;
      }
      ul {
        margin: 0.5rem 0;
        padding-left: 1.1rem;
      }
      .visibility-pick {
        display: grid;
        gap: 0.35rem;
        margin-top: 0.5rem;
        font-size: 0.9rem;
      }
      .vis-label {
        color: #94a3b8;
        font-size: 0.8rem;
        text-transform: uppercase;
        letter-spacing: 0.04em;
      }
      .vis-opt {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        cursor: pointer;
      }
      .adh-badge { font-size: 0.72rem; font-weight: 700; padding: 2px 10px; border-radius: 20px; border: 1px solid; }
      .adh-good { color: #4ade80; border-color: rgba(74,222,128,0.4); background: rgba(74,222,128,0.08); }
      .adh-fair { color: #facc15; border-color: rgba(250,204,21,0.4); background: rgba(250,204,21,0.08); }
      .adh-poor { color: #fb923c; border-color: rgba(251,146,60,0.4); background: rgba(251,146,60,0.08); }
      .adh-very-poor { color: #f87171; border-color: rgba(248,113,113,0.4); background: rgba(248,113,113,0.08); }
      .qa-row { margin-bottom: 0.75rem; }
      .qa-q { font-size: 0.82rem; color: #94a3b8; margin-bottom: 3px; }
      .qa-a { font-size: 0.9rem; color: #e2e8f0; }
      .ai-summary-card h2 { margin-top: 0; }
      .ai-summary-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.75rem;
      }
      .btn-ai {
        padding: 0.4rem 0.9rem;
        background: #4f46e5;
        color: #fff;
        border: none;
        border-radius: 8px;
        cursor: pointer;
        font-size: 0.85rem;
        font-weight: 600;
      }
      .btn-ai:hover:not(:disabled) { background: #4338ca; }
      .btn-ai:disabled { opacity: 0.6; cursor: not-allowed; }
      .ai-summary-body pre {
        white-space: pre-wrap;
        font-family: inherit;
        font-size: 0.9rem;
        line-height: 1.6;
        color: #e2e8f0;
        margin: 0 0 0.5rem;
      }
      .ai-meta { font-size: 0.75rem; color: #64748b; }
      .note-assist-loading { font-size: 0.8rem; color: #94a3b8; margin-top: 0.5rem; font-style: italic; }
      .note-assist-panel {
        margin-top: 0.75rem;
        background: rgba(79, 70, 229, 0.07);
        border: 1px solid rgba(79, 70, 229, 0.3);
        border-radius: 10px;
        padding: 0.75rem 1rem;
      }
      .na-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.6rem; }
      .na-badge {
        font-size: 0.72rem; font-weight: 700; text-transform: uppercase;
        letter-spacing: 0.05em; color: #a5b4fc;
        background: rgba(79,70,229,0.15); padding: 2px 8px; border-radius: 20px;
      }
      .na-provider { font-size: 0.7rem; color: #64748b; }
      .na-section { margin-bottom: 0.6rem; }
      .na-label { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.04em; color: #94a3b8; margin-bottom: 2px; }
      .na-value { font-size: 0.85rem; color: #e2e8f0; }
      .na-soap { white-space: pre-wrap; font-family: inherit; font-size: 0.82rem; color: #cbd5e1; margin: 0; line-height: 1.6; }
      .na-flags { color: #fbbf24; }
      .badge-internal {
        display: inline-block;
        margin-right: 0.35rem;
        padding: 0.1rem 0.45rem;
        border-radius: 6px;
        font-size: 0.72rem;
        font-weight: 600;
        background: rgba(251, 191, 36, 0.15);
        color: #fbbf24;
        border: 1px solid rgba(251, 191, 36, 0.35);
      }
    `
  ]
})
export class DoctorAppointmentDetailComponent implements OnInit, OnDestroy {
  appointment: Appointment | null = null;
  comments: AppointmentComment[] = [];
  commentVisibility: CommentVisibility = 'patient_visible';
  newComment = '';
  posting = false;
  loading = false;
  error = '';

  aiSummary: PatientSummaryResponse | null = null;
  aiLoading = false;
  aiError = '';

  noteAssist: NoteAssistResponse | null = null;
  noteAssistLoading = false;
  private noteInput$ = new Subject<string>();
  private destroy$ = new Subject<void>();

  patientAnswers: { question: string; answer: string }[] = [];
  adherence: { score: number; label: string } | null = null;

  videoLinkInput = '';
  videoLinkSaving = false;
  videoLinkError = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private appointmentsService: AppointmentsService,
    private assistantService: AssistantService,
    private http: HttpClient
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.error = 'Missing appointment id';
      return;
    }
    this.load(id);

    // Debounce note typing → call AI after 1.5s of inactivity
    this.noteInput$.pipe(
      debounceTime(1500),
      distinctUntilChanged(),
      takeUntil(this.destroy$),
      switchMap((text) => {
        this.noteAssistLoading = true;
        return this.assistantService.noteAssist(id, text);
      })
    ).subscribe({
      next: (res) => {
        this.noteAssist = res;
        this.noteAssistLoading = false;
      },
      error: () => { this.noteAssistLoading = false; }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onNoteInput(): void {
    if (this.newComment.trim().length >= 10) {
      this.noteInput$.next(this.newComment);
    }
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
          this.videoLinkInput = appt.video_link ?? '';
          this.reloadComments(id);
          this.loadPatientAnswers(id);
          this.loadAdherence(appt.patient_id);
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
      .postComment(id, body, this.commentVisibility)
      .pipe(finalize(() => (this.posting = false)))
      .subscribe({
        next: () => {
          this.newComment = '';
          this.commentVisibility = 'patient_visible';
          this.reloadComments(id);
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Could not post comment';
        }
      });
  }

  loadAdherence(patientId: string): void {
    this.http.get<{ score: number; label: string }>(ApiContract.patients.adherenceScore(patientId)).subscribe({
      next: (res) => { this.adherence = res; },
      error: () => {}
    });
  }

  loadPatientAnswers(id: string): void {
    this.http.get<{ questions: string; answers: string; submitted_at: string | null }>(
      `${ApiContract.appointments.byId(id)}/questionnaire`
    ).subscribe({
      next: (res) => {
        if (!res.submitted_at || !res.answers) return;
        try {
          const qs: string[] = JSON.parse(res.questions);
          const as_: string[] = JSON.parse(res.answers);
          this.patientAnswers = qs.map((q, i) => ({ question: q, answer: as_[i] ?? '' }));
        } catch {}
      },
      error: () => {}
    });
  }

  loadAiSummary(): void {
    if (!this.appointment) return;
    this.aiLoading = true;
    this.aiError = '';
    this.assistantService.patientSummary(this.appointment.patient_id).subscribe({
      next: (res) => {
        this.aiSummary = res;
        this.aiLoading = false;
      },
      error: (err) => {
        this.aiError = err?.error?.error ?? 'Failed to generate summary';
        this.aiLoading = false;
      }
    });
  }

  saveVideoLink(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) return;
    this.videoLinkSaving = true;
    this.videoLinkError = '';
    this.http.patch(ApiContract.appointments.videoLink(id), { video_link: this.videoLinkInput }).subscribe({
      next: () => {
        if (this.appointment) this.appointment = { ...this.appointment, video_link: this.videoLinkInput || undefined };
        this.videoLinkSaving = false;
      },
      error: (err) => {
        this.videoLinkError = err?.error?.error ?? 'Could not save video link';
        this.videoLinkSaving = false;
      }
    });
  }

  clearVideoLink(): void {
    this.videoLinkInput = '';
    this.saveVideoLink();
  }

  back(): void {
    void this.router.navigate(['/doctor/appointments']);
  }
}

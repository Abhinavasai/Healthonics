import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';

const API = '/api';

interface SoCase {
  id: string;
  title: string;
  anonymized: string;
  specialty: string;
  status: string;
  created_at: string;
  reply_count: number;
}

interface SoReply {
  id: string;
  body: string;
  created_at: string;
  author_label: string;
}

@Component({
  selector: 'app-second-opinion-board',
  standalone: true,
  imports: [CommonModule, FormsModule, DatePipe],
  template: `
    <div class="board">
      <div class="board-header">
        <div>
          <h2>Second-Opinion Board</h2>
          <p class="subtitle">Post anonymized clinical cases and get peer perspectives from other doctors on the platform.</p>
        </div>
        <button class="btn-new" (click)="showForm = !showForm">{{ showForm ? 'Cancel' : '+ New Case' }}</button>
      </div>

      <!-- New Case Form -->
      <div *ngIf="showForm" class="new-case-form card">
        <h3>Post Anonymized Case</h3>
        <div class="form-group">
          <label>Title (what's the clinical question?)</label>
          <input [(ngModel)]="newCase.title" maxlength="200" placeholder="e.g. Persistent fever despite antibiotics — differential?" />
        </div>
        <div class="form-group">
          <label>Specialty</label>
          <select [(ngModel)]="newCase.specialty">
            <option value="">General</option>
            <option value="Cardiology">Cardiology</option>
            <option value="Neurology">Neurology</option>
            <option value="Gastroenterology">Gastroenterology</option>
            <option value="Pulmonology">Pulmonology</option>
            <option value="Infectious Disease">Infectious Disease</option>
            <option value="Endocrinology">Endocrinology</option>
            <option value="Nephrology">Nephrology</option>
            <option value="Rheumatology">Rheumatology</option>
            <option value="Oncology">Oncology</option>
          </select>
        </div>
        <div class="form-group">
          <label>Anonymized Case Description</label>
          <textarea [(ngModel)]="newCase.anonymized" rows="6" maxlength="5000"
            placeholder="Describe the case WITHOUT any identifying information. Include: age range, sex, chief complaint, relevant history, investigations, current treatment, and your specific question."></textarea>
          <div class="char-count">{{ newCase.anonymized.length }}/5000</div>
        </div>
        <div *ngIf="formError" class="error">{{ formError }}</div>
        <button class="btn-post" (click)="createCase()" [disabled]="creating || !newCase.title.trim() || !newCase.anonymized.trim()">
          {{ creating ? 'Posting...' : 'Post Case' }}
        </button>
      </div>

      <!-- Case List -->
      <div *ngIf="loading" class="muted">Loading cases...</div>
      <div *ngIf="!loading && cases.length === 0" class="muted empty">No cases posted yet. Be the first to share a clinical question.</div>

      <div *ngFor="let c of cases" class="case-card card" [class.resolved]="c.status === 'resolved'"
           (click)="selectCase(c)" [class.selected]="selectedCase?.id === c.id">
        <div class="case-top">
          <div class="case-title">{{ c.title }}</div>
          <span *ngIf="c.specialty" class="spec-badge">{{ c.specialty }}</span>
          <span class="status-badge" [class.open]="c.status === 'open'" [class.res]="c.status === 'resolved'">
            {{ c.status === 'resolved' ? 'Resolved' : 'Open' }}
          </span>
        </div>
        <div class="case-preview">{{ c.anonymized | slice:0:180 }}{{ c.anonymized.length > 180 ? '...' : '' }}</div>
        <div class="case-meta">
          <span>{{ c.created_at | date:'MMM d, y' }}</span>
          <span>{{ c.reply_count }} {{ c.reply_count === 1 ? 'reply' : 'replies' }}</span>
        </div>
      </div>

      <!-- Case Detail / Replies Panel -->
      <div *ngIf="selectedCase" class="replies-panel card">
        <div class="replies-header">
          <h3>{{ selectedCase.title }}</h3>
          <button *ngIf="selectedCase.status === 'open'" class="btn-resolve" (click)="resolveCase()">Mark Resolved</button>
        </div>
        <div class="case-full">{{ selectedCase.anonymized }}</div>
        <hr class="divider" />
        <h4>Replies ({{ replies.length }})</h4>
        <div *ngIf="repliesLoading" class="muted">Loading replies...</div>
        <div *ngFor="let r of replies" class="reply">
          <div class="reply-author">{{ r.author_label }}</div>
          <div class="reply-body">{{ r.body }}</div>
          <div class="reply-date">{{ r.created_at | date:'MMM d, y, h:mm a' }}</div>
        </div>
        <div *ngIf="!repliesLoading && replies.length === 0" class="muted">No replies yet. Share your perspective.</div>
        <div *ngIf="selectedCase.status === 'open'" class="reply-form">
          <textarea [(ngModel)]="replyText" rows="3" maxlength="3000"
            placeholder="Share your clinical perspective..." [disabled]="posting"></textarea>
          <button class="btn-reply" (click)="postReply()" [disabled]="posting || !replyText.trim()">
            {{ posting ? 'Posting...' : 'Post Reply' }}
          </button>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .board { max-width: 820px; margin: 0 auto; padding: 16px; }
    .board-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 20px; }
    h2 { margin: 0 0 4px; font-size: 1.5rem; font-weight: 700; }
    .subtitle { color: #94a3b8; font-size: 0.875rem; margin: 0; }
    .btn-new { background: #4f46e5; color: #fff; border: none; border-radius: 8px; padding: 8px 18px; font-weight: 600; cursor: pointer; white-space: nowrap; }
    .btn-new:hover { background: #4338ca; }
    .card { background: rgba(15,23,42,0.7); border: 1px solid rgba(148,163,184,0.2); border-radius: 12px; padding: 16px; margin-bottom: 14px; }
    .new-case-form h3 { margin-top: 0; }
    .form-group { margin-bottom: 12px; }
    label { display: block; font-size: 0.85rem; font-weight: 600; color: #94a3b8; margin-bottom: 5px; }
    input, select, textarea {
      width: 100%; background: rgba(255,255,255,0.05); border: 1px solid rgba(148,163,184,0.25);
      border-radius: 8px; padding: 8px 10px; color: #e2e8f0; font-family: inherit;
      font-size: 0.9rem; resize: vertical; box-sizing: border-box;
    }
    input:focus, select:focus, textarea:focus { outline: none; border-color: #4f46e5; }
    .char-count { text-align: right; font-size: 0.72rem; color: #64748b; margin-top: 3px; }
    .btn-post { background: #4f46e5; color: #fff; border: none; border-radius: 8px; padding: 8px 20px; font-weight: 600; cursor: pointer; }
    .btn-post:disabled { opacity: 0.5; cursor: not-allowed; }
    .error { color: #f87171; font-size: 0.85rem; margin-bottom: 8px; }
    .muted { color: #94a3b8; font-size: 0.875rem; }
    .empty { text-align: center; padding: 32px 0; }
    .case-card { cursor: pointer; transition: border-color 0.15s; }
    .case-card:hover { border-color: rgba(79,70,229,0.5); }
    .case-card.selected { border-color: #4f46e5; }
    .case-card.resolved { opacity: 0.65; }
    .case-top { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 6px; }
    .case-title { font-weight: 600; font-size: 0.95rem; flex: 1; }
    .spec-badge { font-size: 0.72rem; background: rgba(99,102,241,0.15); color: #a5b4fc; border: 1px solid rgba(99,102,241,0.3); border-radius: 20px; padding: 2px 8px; }
    .status-badge { font-size: 0.7rem; font-weight: 700; text-transform: uppercase; padding: 2px 8px; border-radius: 20px; }
    .status-badge.open { background: rgba(74,222,128,0.1); color: #4ade80; border: 1px solid rgba(74,222,128,0.3); }
    .status-badge.res { background: rgba(148,163,184,0.1); color: #94a3b8; border: 1px solid rgba(148,163,184,0.3); }
    .case-preview { font-size: 0.85rem; color: #cbd5e1; line-height: 1.5; margin-bottom: 8px; }
    .case-meta { display: flex; gap: 16px; font-size: 0.75rem; color: #64748b; }
    .replies-panel h3 { margin-top: 0; }
    .replies-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
    .btn-resolve { font-size: 0.78rem; background: transparent; border: 1px solid #4ade80; color: #4ade80; border-radius: 8px; padding: 4px 12px; cursor: pointer; }
    .case-full { font-size: 0.875rem; color: #cbd5e1; white-space: pre-wrap; line-height: 1.6; }
    .divider { border: none; border-top: 1px solid rgba(148,163,184,0.15); margin: 14px 0; }
    h4 { margin: 0 0 10px; font-size: 0.9rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; }
    .reply { border-left: 2px solid rgba(99,102,241,0.4); padding-left: 12px; margin-bottom: 14px; }
    .reply-author { font-size: 0.75rem; font-weight: 700; color: #a5b4fc; margin-bottom: 3px; }
    .reply-body { font-size: 0.875rem; color: #e2e8f0; line-height: 1.5; }
    .reply-date { font-size: 0.72rem; color: #64748b; margin-top: 3px; }
    .reply-form { margin-top: 14px; }
    .btn-reply { margin-top: 8px; background: #4f46e5; color: #fff; border: none; border-radius: 8px; padding: 7px 18px; font-weight: 600; cursor: pointer; }
    .btn-reply:disabled { opacity: 0.5; cursor: not-allowed; }
  `]
})
export class SecondOpinionBoardComponent implements OnInit {
  cases: SoCase[] = [];
  loading = false;
  showForm = false;
  creating = false;
  formError = '';

  newCase = { title: '', anonymized: '', specialty: '' };

  selectedCase: SoCase | null = null;
  replies: SoReply[] = [];
  repliesLoading = false;
  replyText = '';
  posting = false;

  constructor(private http: HttpClient) {}

  ngOnInit(): void { this.load(); }

  load(): void {
    this.loading = true;
    this.http.get<{ cases: SoCase[] }>(`${API}/second-opinions`).subscribe({
      next: (res) => { this.cases = res.cases ?? []; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  createCase(): void {
    this.creating = true;
    this.formError = '';
    this.http.post(`${API}/second-opinions`, this.newCase).subscribe({
      next: () => {
        this.creating = false;
        this.showForm = false;
        this.newCase = { title: '', anonymized: '', specialty: '' };
        this.load();
      },
      error: (err) => {
        this.formError = err?.error?.error ?? 'Could not post case';
        this.creating = false;
      }
    });
  }

  selectCase(c: SoCase): void {
    this.selectedCase = c;
    this.replyText = '';
    this.loadReplies(c.id);
  }

  loadReplies(caseId: string): void {
    this.repliesLoading = true;
    this.http.get<{ replies: SoReply[] }>(`${API}/second-opinions/${caseId}/replies`).subscribe({
      next: (res) => { this.replies = res.replies ?? []; this.repliesLoading = false; },
      error: () => { this.repliesLoading = false; }
    });
  }

  postReply(): void {
    if (!this.selectedCase) return;
    this.posting = true;
    this.http.post(`${API}/second-opinions/${this.selectedCase.id}/replies`, { body: this.replyText }).subscribe({
      next: () => {
        this.replyText = '';
        this.posting = false;
        this.loadReplies(this.selectedCase!.id);
        if (this.selectedCase) this.selectedCase.reply_count++;
      },
      error: () => { this.posting = false; }
    });
  }

  resolveCase(): void {
    if (!this.selectedCase) return;
    this.http.patch(`${API}/second-opinions/${this.selectedCase.id}/resolve`, {}).subscribe({
      next: () => {
        if (this.selectedCase) this.selectedCase.status = 'resolved';
        this.load();
      },
      error: () => {}
    });
  }
}

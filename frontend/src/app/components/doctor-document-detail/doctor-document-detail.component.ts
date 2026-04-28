/** Doctor document detail + AI summary (Sprint 3 F4) — lane s3.5-abhinav. */
import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { Subscription, timer } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { DoctorDocumentDetail, DoctorDocumentsService } from '../../services/doctor-documents.service';
import { ContextualCommentsPanelComponent } from '../contextual-comments-panel/contextual-comments-panel.component';

@Component({
  selector: 'app-doctor-document-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe, ContextualCommentsPanelComponent],
  templateUrl: './doctor-document-detail.component.html',
  styleUrl: './doctor-document-detail.component.scss'
})
export class DoctorDocumentDetailComponent implements OnInit, OnDestroy {
  doc: DoctorDocumentDetail | null = null;
  loading = true;
  summarizeLoading = false;
  error = '';
  /** Last error from POST /summarize (shown inline; distinct from page load error). */
  summaryRequestError = '';
  summarizeInfo = '';
  showComments = false;
  private documentId = '';
  private pollSub?: Subscription;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private api: DoctorDocumentsService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe((params) => {
      const id = params.get('documentId');
      if (!id) {
        void this.router.navigate(['/doctor/documents']);
        return;
      }
      this.documentId = id;
      this.stopPoll();
      this.loadInitial();
    });
  }

  ngOnDestroy(): void {
    this.stopPoll();
  }

  private loadInitial(): void {
    this.loading = true;
    this.error = '';
    this.summaryRequestError = '';
    this.summarizeInfo = '';
    this.doc = null;
    this.api.getDetail(this.documentId).subscribe({
      next: (d) => {
        this.doc = d;
        this.loading = false;
        if (!d) {
          this.error = 'Document not found or you do not have access.';
          return;
        }
        if (d.summary_status === 'pending') {
          this.startPoll();
        }
      },
      error: (err) => {
        this.doc = null;
        this.loading = false;
        this.error = err?.error?.error ?? 'Unable to load document.';
      }
    });
  }

  private startPoll(): void {
    this.stopPoll();
    this.pollSub = timer(3000, 3000).subscribe(() => {
      if (this.doc?.summary_status !== 'pending') {
        this.stopPoll();
        return;
      }
      this.api.getDetail(this.documentId).subscribe({
        next: (d) => {
          if (d) {
            this.doc = d;
            if (d.summary_status !== 'pending') {
              this.stopPoll();
            }
          }
        }
      });
    });
  }

  private stopPoll(): void {
    this.pollSub?.unsubscribe();
    this.pollSub = undefined;
  }

  /** Re-run the async summary pipeline (generate, regenerate, or retry after failure). */
  requestSummary(): void {
    if (!this.documentId || this.summarizeLoading) {
      return;
    }
    this.summarizeLoading = true;
    this.error = '';
    this.summaryRequestError = '';
    this.summarizeInfo = '';
    this.api
      .requestSummary(this.documentId)
      .pipe(finalize(() => (this.summarizeLoading = false)))
      .subscribe({
        next: (res) => {
          if (res === null) {
            this.summaryRequestError = 'Could not start summarization. Please try again.';
            return;
          }
          this.summaryRequestError = '';
          if (res.message) {
            this.summarizeInfo = res.message;
          } else if (res.job_id) {
            this.summarizeInfo = 'Summary job queued. Refreshing status...';
          }
          this.api.getDetail(this.documentId).subscribe({
            next: (d) => {
              this.doc = d;
              if (d?.summary_status === 'pending') {
                this.startPoll();
              }
            }
          });
        },
        error: (err) => {
          this.summaryRequestError = err?.error?.error ?? 'Could not start summarization. Check your connection and try again.';
        }
      });
  }

  /** Alias for template clarity on failed state. */
  retrySummary(): void {
    this.requestSummary();
  }

  formatBytes(n: number): string {
    if (n < 1024) {
      return `${n} B`;
    }
    if (n < 1024 * 1024) {
      return `${(n / 1024).toFixed(1)} KB`;
    }
    return `${(n / (1024 * 1024)).toFixed(1)} MB`;
  }

  showSummaryBlock(): boolean {
    return !!this.doc && (this.doc.summary_status === 'ready' || !!this.doc.summary);
  }

  statusLabel(s?: string): string {
    switch (s) {
      case 'pending':
        return 'Summary pending';
      case 'ready':
        return 'Summary ready';
      case 'failed':
        return 'Summary failed';
      default:
        return 'No summary yet';
    }
  }

  statusClass(s?: string): string {
    switch (s) {
      case 'pending':
        return 'pill pill-pending';
      case 'ready':
        return 'pill pill-ready';
      case 'failed':
        return 'pill pill-failed';
      default:
        return 'pill';
    }
  }

  canInlinePreview(d: DoctorDocumentDetail): boolean {
    const ct = (d.content_type || '').toLowerCase();
    return ct === 'application/pdf' || ct.startsWith('image/');
  }

  previewCaption(d: DoctorDocumentDetail): string {
    if (this.canInlinePreview(d)) {
      return 'Inline preview is enabled for this file type. Secure viewer endpoint will stream content in the next PR.';
    }
    return 'Inline preview is unavailable for this file type. Open/download support will use secure access flow.';
  }

  /** Primary CTA label for the summary action button. */
  primarySummaryLabel(d: DoctorDocumentDetail): string {
    if (this.summarizeLoading) {
      return 'Starting…';
    }
    switch (d.summary_status) {
      case 'pending':
        return 'Generating…';
      case 'ready':
        return 'Regenerate summary';
      case 'failed':
        return 'Try again';
      default:
        return 'Generate summary';
    }
  }

  primarySummaryDisabled(d: DoctorDocumentDetail): boolean {
    return this.summarizeLoading || d.summary_status === 'pending';
  }

  toggleComments(): void {
    this.showComments = !this.showComments;
  }
}

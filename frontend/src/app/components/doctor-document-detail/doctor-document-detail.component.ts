import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { Subscription, timer } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { DoctorDocumentDetail, DoctorDocumentsService } from '../../services/doctor-documents.service';

@Component({
  selector: 'app-doctor-document-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe],
  templateUrl: './doctor-document-detail.component.html',
  styleUrl: './doctor-document-detail.component.scss'
})
export class DoctorDocumentDetailComponent implements OnInit, OnDestroy {
  doc: DoctorDocumentDetail | null = null;
  loading = true;
  summarizeLoading = false;
  error = '';
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

  requestSummary(): void {
    if (!this.documentId || this.summarizeLoading) {
      return;
    }
    this.summarizeLoading = true;
    this.error = '';
    this.api
      .requestSummary(this.documentId)
      .pipe(finalize(() => (this.summarizeLoading = false)))
      .subscribe({
        next: (res) => {
          if (res === null) {
            this.error = 'Could not start summarization.';
            return;
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
          this.error = err?.error?.error ?? 'Could not start summarization.';
        }
      });
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
}

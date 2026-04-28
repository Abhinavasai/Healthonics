import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import {
  KnowledgeAdminService,
  KnowledgeDetail,
  KnowledgeListDocument,
  SimilarityPair
} from '../../services/knowledge-admin.service';
import { ContextualCommentsPanelComponent } from '../contextual-comments-panel/contextual-comments-panel.component';

@Component({
  selector: 'app-admin-knowledge',
  standalone: true,
  imports: [CommonModule, FormsModule, ContextualCommentsPanelComponent],
  templateUrl: './admin-knowledge.component.html',
  styleUrl: './admin-knowledge.component.scss'
})
export class AdminKnowledgeComponent implements OnInit {
  docs: KnowledgeListDocument[] = [];
  title = '';
  body = '';
  reviewIntervalDays: number | null = null;
  saving = false;
  loading = false;
  error: string | null = null;

  selectedId: string | null = null;
  detail: KnowledgeDetail | null = null;
  detailLoading = false;
  patchTitle = '';
  patchBody = '';
  patchReviewDays: number | null = null;
  patching = false;
  reviewing = false;

  versionsOpen = false;
  versions: { version: number; excerpt: string; created_at: string }[] = [];
  versionsLoading = false;

  scanThreshold = 0.85;
  scanLoading = false;
  scanPairs: SimilarityPair[] = [];
  scanMeta: { chunks_scanned: number; threshold: number } | null = null;
  showDetailComments = false;

  constructor(private kb: KnowledgeAdminService) {}

  ngOnInit(): void {
    this.reload();
  }

  get analytics(): {
    total: number;
    stale: number;
    avgHealth: number | null;
    needsAttention: number;
  } {
    const d = this.docs;
    const withHealth = d.filter((x) => x.health_score != null);
    const stale = d.filter((x) => x.is_stale).length;
    const avg =
      withHealth.length > 0
        ? Math.round(
            withHealth.reduce((s, x) => s + (x.health_score ?? 0), 0) / withHealth.length
          )
        : null;
    const needsAttention = d.filter(
      (x) => (x.health_score ?? 100) < 70 || x.is_stale
    ).length;
    return { total: d.length, stale, avgHealth: avg, needsAttention };
  }

  reload(): void {
    this.loading = true;
    this.error = null;
    this.kb
      .listDocs()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (r) => (this.docs = r.documents ?? []),
        error: () => (this.error = 'Could not load documents')
      });
  }

  create(): void {
    this.saving = true;
    this.error = null;
    const payload: { title: string; body: string; review_interval_days?: number } = {
      title: this.title.trim(),
      body: this.body.trim()
    };
    if (this.reviewIntervalDays != null && this.reviewIntervalDays >= 1 && this.reviewIntervalDays <= 3650) {
      payload.review_interval_days = this.reviewIntervalDays;
    }
    this.kb.createDoc(payload).subscribe({
      next: () => {
        this.title = '';
        this.body = '';
        this.reviewIntervalDays = null;
        this.saving = false;
        this.reload();
      },
      error: (err) => {
        this.saving = false;
        this.error = err?.error?.error ?? 'Save failed';
      }
    });
  }

  selectDoc(doc: KnowledgeListDocument): void {
    if (this.selectedId === doc.id) {
      this.clearSelection();
      return;
    }
    this.selectedId = doc.id;
    this.detail = null;
    this.versions = [];
    this.versionsOpen = false;
    this.detailLoading = true;
    this.kb.getDoc(doc.id).subscribe({
      next: (d) => {
        this.detail = d;
        this.patchTitle = d.title;
        this.patchBody = d.body;
        this.patchReviewDays = d.review_interval_days;
        this.detailLoading = false;
      },
      error: () => {
        this.detailLoading = false;
        this.error = 'Could not load document';
      }
    });
  }

  clearSelection(): void {
    this.selectedId = null;
    this.detail = null;
    this.versions = [];
    this.versionsOpen = false;
    this.showDetailComments = false;
  }

  savePatch(): void {
    if (!this.selectedId || !this.detail) {
      return;
    }
    this.patching = true;
    this.error = null;
    const payload: Partial<{ title: string; body: string; review_interval_days: number }> = {};
    const t = this.patchTitle.trim();
    const b = this.patchBody.trim();
    if (t && t !== this.detail.title) {
      payload.title = t;
    }
    if (b && b !== this.detail.body) {
      payload.body = b;
    }
    const rd = this.patchReviewDays;
    if (rd != null && rd >= 1 && rd <= 3650 && rd !== this.detail.review_interval_days) {
      payload.review_interval_days = rd;
    }
    if (Object.keys(payload).length === 0) {
      this.patching = false;
      return;
    }
    this.kb.patchDoc(this.selectedId, payload).subscribe({
      next: () => {
        this.patching = false;
        this.reload();
        this.kb.getDoc(this.selectedId!).subscribe((d) => {
          this.detail = d;
        });
      },
      error: (err) => {
        this.patching = false;
        this.error = err?.error?.error ?? 'Update failed';
      }
    });
  }

  review(): void {
    if (!this.selectedId) {
      return;
    }
    this.reviewing = true;
    this.kb.markReviewed(this.selectedId).subscribe({
      next: () => {
        this.reviewing = false;
        this.reload();
        if (this.selectedId) {
          this.kb.getDoc(this.selectedId).subscribe((d) => (this.detail = d));
        }
      },
      error: (err) => {
        this.reviewing = false;
        this.error = err?.error?.error ?? 'Review failed';
      }
    });
  }

  toggleVersions(): void {
    if (!this.selectedId) {
      return;
    }
    this.versionsOpen = !this.versionsOpen;
    if (!this.versionsOpen || this.versions.length) {
      return;
    }
    this.versionsLoading = true;
    this.kb.listVersions(this.selectedId).subscribe({
      next: (r) => {
        this.versions = (r.versions ?? []).map((v) => ({
          version: v.version,
          excerpt: v.excerpt,
          created_at: v.created_at
        }));
        this.versionsLoading = false;
      },
      error: () => {
        this.versionsLoading = false;
        this.error = 'Could not load versions';
      }
    });
  }

  runSimilarityScan(): void {
    this.scanLoading = true;
    this.scanPairs = [];
    this.scanMeta = null;
    this.kb.similarityScan(this.scanThreshold, 1200).subscribe({
      next: (r) => {
        this.scanPairs = r.pairs ?? [];
        this.scanMeta = {
          chunks_scanned: r.chunks_scanned,
          threshold: r.threshold
        };
        this.scanLoading = false;
      },
      error: () => {
        this.scanLoading = false;
        this.error = 'Similarity scan failed (ensure embeddings are uploaded for KB chunks).';
      }
    });
  }

  healthTone(score?: number): string {
    if (score == null) {
      return 'muted';
    }
    if (score >= 75) {
      return 'good';
    }
    if (score >= 45) {
      return 'warn';
    }
    return 'bad';
  }

  similarityBadgeClass(hint: string): string {
    if (hint.includes('conflict')) {
      return 'badge--danger';
    }
    if (hint.includes('duplicate')) {
      return 'badge--purple';
    }
    return 'badge--info';
  }

  toggleDetailComments(): void {
    this.showDetailComments = !this.showDetailComments;
  }
}

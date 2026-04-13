import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { finalize } from 'rxjs';
import { PatientDocumentRow, PatientDocumentsService } from '../../services/patient-documents.service';

@Component({
  selector: 'app-patient-documents',
  standalone: true,
  imports: [CommonModule, DatePipe],
  template: `
    <section class="docs" data-cy="patient-documents-page">
      <h1>My Documents</h1>
      <p class="subtitle">Upload your reports and manage downloaded copies.</p>

      <div class="card upload">
        <input type="file" (change)="onFilePicked($event)" [disabled]="uploading" />
        <button type="button" (click)="upload()" [disabled]="!pendingFile || uploading">
          {{ uploading ? 'Uploading...' : 'Upload' }}
        </button>
        <p class="muted" *ngIf="pendingFile">{{ pendingFile.name }}</p>
      </div>

      <p *ngIf="message" class="msg">{{ message }}</p>
      <p *ngIf="error" class="error">{{ error }}</p>

      <div class="card">
        <div class="head">
          <h2>Uploaded files</h2>
          <button type="button" (click)="load()" [disabled]="loading">{{ loading ? 'Refreshing...' : 'Refresh' }}</button>
        </div>
        <p *ngIf="!loading && !documents.length" class="muted">No documents uploaded yet.</p>
        <ul *ngIf="documents.length">
          <li *ngFor="let d of documents">
            <div>
              <strong>{{ d.filename }}</strong>
              <p>{{ d.created_at | date: 'medium' }} · {{ prettyBytes(d.size_bytes) }}</p>
            </div>
            <a [href]="downloadHref(d.id)">Download</a>
          </li>
        </ul>
      </div>
    </section>
  `,
  styles: [`
    .docs { max-width: 860px; margin: 0 auto; display: grid; gap: 1rem; }
    .subtitle,.muted { color: #94a3b8; }
    .card { background: rgba(15, 23, 42, .7); border: 1px solid rgba(148, 163, 184, .2); border-radius: 12px; padding: 1rem; }
    .upload { display: flex; flex-wrap: wrap; gap: .65rem; align-items: center; }
    .head { display: flex; justify-content: space-between; align-items: center; margin-bottom: .6rem; }
    button,a { width: fit-content; padding: .5rem .8rem; border-radius: 8px; border: 1px solid #22d3ee; background: #0f172a; color: #22d3ee; text-decoration: none; cursor: pointer; }
    ul { list-style: none; margin: 0; padding: 0; display: grid; gap: .55rem; }
    li { border: 1px solid rgba(148,163,184,.2); border-radius: 8px; padding: .65rem; display: flex; align-items: center; justify-content: space-between; gap: .75rem; }
    p { margin: .2rem 0 0; }
    .error { color: #f87171; }
    .msg { color: #22d3ee; }
  `]
})
export class PatientDocumentsComponent implements OnInit {
  loading = false;
  uploading = false;
  documents: PatientDocumentRow[] = [];
  pendingFile: File | null = null;
  message = '';
  error = '';

  constructor(private api: PatientDocumentsService) {}

  ngOnInit(): void {
    this.load();
  }

  onFilePicked(evt: Event): void {
    const input = evt.target as HTMLInputElement;
    this.pendingFile = input.files?.[0] ?? null;
  }

  upload(): void {
    if (!this.pendingFile) {
      return;
    }
    this.error = '';
    this.message = '';
    this.uploading = true;
    this.api.upload(this.pendingFile).pipe(finalize(() => (this.uploading = false))).subscribe({
      next: () => {
        this.message = 'Document uploaded.';
        this.pendingFile = null;
        this.load();
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Upload failed.';
      }
    });
  }

  load(): void {
    this.loading = true;
    this.api.list().pipe(finalize(() => (this.loading = false))).subscribe({
      next: (res) => (this.documents = res.documents ?? []),
      error: (err) => {
        this.documents = [];
        this.error = err?.error?.error ?? 'Could not load documents.';
      }
    });
  }

  downloadHref(id: string): string {
    return this.api.downloadUrl(id);
  }

  prettyBytes(bytes: number): string {
    if (!bytes || bytes < 1024) return `${bytes || 0} B`;
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
}


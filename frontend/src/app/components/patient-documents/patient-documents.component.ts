import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { finalize, Subscription } from 'rxjs';
import { DocumentsService, PatientDocument } from '../../services/documents.service';

@Component({
  selector: 'app-patient-documents',
  standalone: true,
  imports: [CommonModule, DatePipe],
  templateUrl: './patient-documents.component.html',
  styleUrl: './patient-documents.component.scss'
})
export class PatientDocumentsComponent implements OnInit, OnDestroy {
  documents: PatientDocument[] = [];
  loading = false;
  uploading = false;
  uploadPercent: number | null = null;
  error = '';
  formMessage = '';
  private uploadSub?: Subscription;

  constructor(private documentsApi: DocumentsService) {}

  ngOnInit(): void {
    this.load();
  }

  ngOnDestroy(): void {
    this.uploadSub?.unsubscribe();
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.documentsApi
      .list()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (list) => (this.documents = list),
        error: (err) => {
          this.documents = [];
          this.error = err?.error?.error ?? 'Unable to load documents.';
        }
      });
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) {
      return;
    }
    this.uploadFile(file);
  }

  uploadFile(file: File): void {
    this.uploadSub?.unsubscribe();
    this.uploading = true;
    this.uploadPercent = 0;
    this.formMessage = '';
    this.error = '';

    this.uploadSub = this.documentsApi.upload(file).subscribe({
      next: (ev) => {
        if (ev.type === 'progress') {
          this.uploadPercent = ev.percent;
        } else if (ev.type === 'done') {
          this.formMessage = `Uploaded “${ev.doc.filename}”.`;
          this.load();
        } else if (ev.type === 'error') {
          this.error = ev.message;
        }
      },
      error: () => {
        this.error = 'Upload failed.';
      },
      complete: () => {
        this.uploading = false;
        this.uploadPercent = null;
      }
    });
  }

  openInNewTab(doc: PatientDocument): void {
    this.documentsApi.downloadBlob(doc.id).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        window.open(url, '_blank', 'noopener,noreferrer');
        setTimeout(() => URL.revokeObjectURL(url), 120_000);
      },
      error: (err) => {
        this.error = err?.error?.error ?? 'Unable to download file.';
      }
    });
  }

  formatBytes(n?: number): string {
    if (n == null || n < 0) {
      return '—';
    }
    if (n < 1024) {
      return `${n} B`;
    }
    if (n < 1024 * 1024) {
      return `${(n / 1024).toFixed(1)} KB`;
    }
    return `${(n / (1024 * 1024)).toFixed(1)} MB`;
  }

  statusLabel(status?: string): string {
    switch (status) {
      case 'pending':
        return 'Processing';
      case 'ready':
        return 'Ready';
      case 'failed':
        return 'Failed';
      default:
        return 'Ready';
    }
  }
}

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { PatientFilesService, PatientFileRow } from '../../services/patient-files.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-patient-files',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="wrap" data-cy="patient-files-page">
      <h1>My files</h1>
      <p class="hint">Upload PDF or images up to 5 MB.</p>
      <div class="upload" data-cy="patient-files-upload">
        <input type="file" accept=".pdf,.png,.jpg,.jpeg" #fileIn data-cy="patient-files-input" />
        <input
          type="text"
          placeholder="Description (optional)"
          [(ngModel)]="description"
          data-cy="patient-files-description"
        />
        <button type="button" (click)="upload(fileIn)" [disabled]="busy" data-cy="patient-files-submit">
          Upload
        </button>
      </div>
      <p class="err" *ngIf="error">{{ error }}</p>
      <ul class="list" *ngIf="files.length" data-cy="patient-files-list">
        <li *ngFor="let f of files" data-cy="patient-files-row">
          <span>{{ f.original_name }}</span>
          <small>({{ f.byte_size }} bytes)</small>
          <button type="button" (click)="download(f)" data-cy="patient-files-download">Download</button>
        </li>
      </ul>
      <p *ngIf="!busy && !error && !files.length" data-cy="patient-files-empty">No files yet.</p>
    </section>
  `,
  styles: [
    `
      .wrap {
        max-width: 640px;
      }
      .upload {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        margin-bottom: 1rem;
        align-items: center;
      }
      .err {
        color: #b00020;
      }
      .list {
        list-style: none;
        padding: 0;
      }
      .list li {
        padding: 0.5rem 0;
        border-bottom: 1px solid #eee;
        display: flex;
        gap: 0.75rem;
        align-items: baseline;
        flex-wrap: wrap;
      }
      .hint {
        color: #555;
      }
    `
  ]
})
export class PatientFilesComponent implements OnInit {
  files: PatientFileRow[] = [];
  description = '';
  error: string | null = null;
  busy = false;

  constructor(
    private api: PatientFilesService,
    private auth: AuthService
  ) {}

  ngOnInit(): void {
    this.reload();
  }

  reload(): void {
    const u = this.auth.getUser();
    if (!u) {
      this.error = 'Not signed in';
      return;
    }
    this.busy = true;
    this.error = null;
    this.api.list(u.id).subscribe({
      next: (res) => {
        this.files = res.files ?? [];
        this.busy = false;
      },
      error: () => {
        this.error = 'Could not load files';
        this.busy = false;
      }
    });
  }

  upload(fileIn: HTMLInputElement): void {
    const u = this.auth.getUser();
    if (!u || !fileIn.files?.length) {
      this.error = 'Choose a file';
      return;
    }
    const f = fileIn.files[0];
    this.busy = true;
    this.error = null;
    this.api.upload(u.id, f, this.description).subscribe({
      next: () => {
        this.description = '';
        fileIn.value = '';
        this.reload();
      },
      error: () => {
        this.error = 'Upload failed';
        this.busy = false;
      }
    });
  }

  download(f: PatientFileRow): void {
    this.api.download(f.id).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = f.original_name || 'download';
        a.click();
        URL.revokeObjectURL(url);
      },
      error: () => {
        this.error = 'Download failed';
      }
    });
  }
}

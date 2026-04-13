import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { finalize } from 'rxjs';
import { PatientDocumentsService } from '../../services/patient-documents.service';

@Component({
  selector: 'app-patient-documents-upload',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './patient-documents-upload.component.html',
  styleUrl: './patient-documents-upload.component.scss'
})
export class PatientDocumentsUploadComponent {
  selectedFile: File | null = null;
  loading = false;
  error = '';
  success = '';

  constructor(private api: PatientDocumentsService) {}

  onFileChange(ev: Event): void {
    const input = ev.target as HTMLInputElement;
    const f = input.files?.[0];
    this.selectedFile = f ?? null;
    this.error = '';
    this.success = '';
  }

  submit(): void {
    if (!this.selectedFile || this.loading) {
      return;
    }
    this.loading = true;
    this.error = '';
    this.success = '';
    this.api
      .upload(this.selectedFile)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (res) => {
          this.success = `Uploaded “${res.filename}”. Your care team can open it under Patient documents after you have an appointment with them.`;
          this.selectedFile = null;
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Upload failed.';
        }
      });
  }
}

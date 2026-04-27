/** Doctor documents inbox (Sprint 3 F4) — lane s3.5-karthik. */
import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterModule } from '@angular/router';
import { finalize } from 'rxjs';
import { DoctorDocumentsService, DoctorDocumentListItem } from '../../services/doctor-documents.service';

@Component({
  selector: 'app-doctor-documents-list',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe],
  templateUrl: './doctor-documents-list.component.html',
  styleUrl: './doctor-documents-list.component.scss'
})
export class DoctorDocumentsListComponent implements OnInit {
  documents: DoctorDocumentListItem[] = [];
  loading = false;
  error = '';

  constructor(private api: DoctorDocumentsService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.api
      .list()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: (rows) => (this.documents = rows),
        error: (err) => {
          this.documents = [];
          this.error = err?.error?.error ?? 'Unable to load patient documents.';
        }
      });
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
        return '—';
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
}

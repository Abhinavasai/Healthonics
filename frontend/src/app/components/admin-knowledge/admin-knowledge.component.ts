import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';

interface KnowledgeRow {
  id: string;
  title: string;
  excerpt: string;
  created_at: string;
  updated_at: string;
}

@Component({
  selector: 'app-admin-knowledge',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section data-cy="admin-knowledge-page">
      <h1>Knowledge base</h1>
      <p *ngIf="error" class="err">{{ error }}</p>
      <form class="form" (ngSubmit)="create()">
        <h2>New document</h2>
        <label>Title <input [(ngModel)]="title" name="kt" required data-cy="knowledge-title" /></label>
        <label>Body <textarea [(ngModel)]="body" name="kb" rows="4" required data-cy="knowledge-body"></textarea></label>
        <button type="submit" [disabled]="saving || !title.trim() || !body.trim()" data-cy="knowledge-save">Save</button>
      </form>
      <h2>Documents</h2>
      <ul *ngIf="docs.length" data-cy="knowledge-doc-list">
        <li *ngFor="let d of docs" data-cy="knowledge-doc-row">{{ d.title }}</li>
      </ul>
      <p *ngIf="!error && !docs.length" data-cy="knowledge-empty">No documents yet.</p>
    </section>
  `,
  styles: [
    `
      .form {
        display: grid;
        gap: 0.5rem;
        margin-bottom: 1.5rem;
        max-width: 520px;
      }
      label {
        display: grid;
        gap: 0.25rem;
      }
      input,
      textarea {
        background: #0b1220;
        color: #e5e7eb;
        border: 1px solid #334155;
        border-radius: 8px;
        padding: 0.45rem;
      }
      button {
        width: fit-content;
        padding: 0.45rem 0.75rem;
        border-radius: 8px;
        border: 1px solid #22d3ee;
        background: #0f172a;
        color: #22d3ee;
        cursor: pointer;
      }
      .err {
        color: #f87171;
      }
    `
  ]
})
export class AdminKnowledgeComponent implements OnInit {
  docs: KnowledgeRow[] = [];
  title = '';
  body = '';
  saving = false;
  error: string | null = null;

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.reload();
  }

  reload(): void {
    this.http.get<{ documents: KnowledgeRow[] }>('/api/admin/knowledge-docs').subscribe({
      next: (r) => (this.docs = r.documents ?? []),
      error: () => (this.error = 'Could not load documents')
    });
  }

  create(): void {
    this.saving = true;
    this.error = null;
    this.http
      .post('/api/admin/knowledge-docs', { title: this.title.trim(), body: this.body.trim() })
      .subscribe({
        next: () => {
          this.title = '';
          this.body = '';
          this.saving = false;
          this.reload();
        },
        error: (err) => {
          this.saving = false;
          this.error = err?.error?.error ?? 'Save failed';
        }
      });
  }
}

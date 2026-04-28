import { CommonModule } from '@angular/common';
import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';
import {
  CommentVisibility,
  ContextType,
  ContextualCommentRow,
  ContextualCommentsService
} from '../../services/contextual-comments.service';

@Component({
  selector: 'app-contextual-comments-panel',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <section class="panel" data-cy="context-comments-panel">
      <h3>{{ panelTitle }}</h3>
      <p class="err" *ngIf="error">{{ error }}</p>

      <div class="composer">
        <textarea
          rows="3"
          name="commentBody"
          [(ngModel)]="newBody"
          placeholder="Add comment"
          data-cy="context-comments-body"
        ></textarea>
        <div class="composer-row">
          <select
            name="commentVisibility"
            [(ngModel)]="newVisibility"
            [disabled]="visibilityOptions.length <= 1"
            data-cy="context-comments-visibility"
          >
            <option *ngFor="let v of visibilityOptions" [value]="v">{{ v }}</option>
          </select>
          <button type="button" (click)="create()" [disabled]="saving || !newBody.trim()" data-cy="context-comments-add">
            {{ saving ? 'Posting...' : 'Add comment' }}
          </button>
        </div>
      </div>

      <p *ngIf="loading">Loading comments...</p>
      <ul *ngIf="!loading && comments.length" class="comment-list" data-cy="context-comments-list">
        <li *ngFor="let c of comments" data-cy="context-comment-row">
          <div class="meta">
            <span>{{ c.visibility }}</span>
            <small>{{ c.created_at }}</small>
          </div>
          <p>{{ c.body }}</p>
        </li>
      </ul>
      <p *ngIf="!loading && !comments.length" class="muted">No comments yet.</p>
    </section>
  `,
  styles: [
    `
      .panel {
        border: 1px solid rgba(148, 163, 184, 0.28);
        border-radius: 0.5rem;
        padding: 0.75rem;
        margin-top: 0.75rem;
      }
      .composer {
        display: grid;
        gap: 0.4rem;
        margin-bottom: 0.65rem;
      }
      textarea,
      select {
        background: rgba(15, 23, 42, 0.65);
        border: 1px solid rgba(148, 163, 184, 0.35);
        border-radius: 0.35rem;
        color: #e2e8f0;
        padding: 0.45rem 0.55rem;
      }
      .composer-row {
        display: flex;
        gap: 0.5rem;
        justify-content: space-between;
      }
      .comment-list {
        list-style: none;
        margin: 0;
        padding: 0;
        display: grid;
        gap: 0.45rem;
      }
      .comment-list li {
        border: 1px solid rgba(148, 163, 184, 0.25);
        border-radius: 0.4rem;
        padding: 0.45rem 0.55rem;
      }
      .meta {
        display: flex;
        justify-content: space-between;
        color: #94a3b8;
        font-size: 0.8rem;
      }
      .err {
        color: #f87171;
      }
      .muted {
        color: #94a3b8;
      }
    `
  ]
})
export class ContextualCommentsPanelComponent implements OnChanges {
  @Input({ required: true }) contextType!: ContextType;
  @Input({ required: true }) contextId!: string;
  @Input() panelTitle = 'Comments';

  loading = false;
  saving = false;
  error: string | null = null;
  comments: ContextualCommentRow[] = [];
  newBody = '';
  newVisibility: CommentVisibility = 'patient_visible';

  constructor(
    private api: ContextualCommentsService,
    private auth: AuthService
  ) {}

  get visibilityOptions(): CommentVisibility[] {
    const role = this.auth.getUser()?.role ?? '';
    if (role === 'patient') {
      return ['patient_visible'];
    }
    return ['patient_visible', 'care_team', 'internal'];
  }

  ngOnChanges(changes: SimpleChanges): void {
    if ((changes['contextType'] || changes['contextId']) && this.contextType && this.contextId) {
      this.newVisibility = this.visibilityOptions[0];
      this.load();
    }
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.list(this.contextType, this.contextId).subscribe({
      next: (res) => {
        this.comments = res.comments ?? [];
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.error = 'Could not load comments.';
      }
    });
  }

  create(): void {
    this.saving = true;
    this.error = null;
    this.api
      .create(this.contextType, this.contextId, {
        body: this.newBody.trim(),
        visibility: this.newVisibility
      })
      .subscribe({
        next: () => {
          this.newBody = '';
          this.saving = false;
          this.load();
        },
        error: () => {
          this.saving = false;
          this.error = 'Could not post comment.';
        }
      });
  }
}


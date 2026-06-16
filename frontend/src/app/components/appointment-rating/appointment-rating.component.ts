import { Component, Input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { ApiContract } from '../../services/api-contract';

@Component({
  selector: 'app-appointment-rating',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="rating-card">
      <h3>Rate your visit</h3>
      <div *ngIf="submitted" class="success">Thank you for your feedback!</div>
      <div *ngIf="existing" class="existing">
        You rated this visit <strong>{{ existing.rating }}/5</strong>
        <span *ngIf="existing.comment"> — "{{ existing.comment }}"</span>
      </div>
      <div *ngIf="!submitted && !existing">
        <div class="stars">
          <button *ngFor="let s of [1,2,3,4,5]" type="button"
            class="star" [class.active]="selectedRating >= s"
            (click)="selectedRating = s">★</button>
        </div>
        <textarea [(ngModel)]="comment" rows="2" placeholder="Optional comment..." class="comment-input"></textarea>
        <button type="button" (click)="submit()" [disabled]="!selectedRating || saving">
          {{ saving ? 'Submitting...' : 'Submit Rating' }}
        </button>
        <p *ngIf="error" class="error">{{ error }}</p>
      </div>
    </div>
  `,
  styles: [`
    .rating-card { background: rgba(15,23,42,0.7); border: 1px solid rgba(148,163,184,0.2); border-radius: 12px; padding: 1rem; }
    h3 { margin: 0 0 0.75rem; font-size: 1rem; }
    .stars { display: flex; gap: 0.25rem; margin-bottom: 0.75rem; }
    .star { background: none; border: none; font-size: 2rem; cursor: pointer; color: #475569; transition: color 0.15s; }
    .star.active { color: #fbbf24; }
    .comment-input { width: 100%; box-sizing: border-box; background: rgba(255,255,255,0.05); border: 1px solid rgba(148,163,184,0.25); border-radius: 8px; padding: 8px 10px; color: #e2e8f0; font-family: inherit; font-size: 0.88rem; resize: vertical; margin-bottom: 0.5rem; }
    .success { color: #4ade80; font-size: 0.9rem; }
    .existing { color: #94a3b8; font-size: 0.9rem; }
    .error { color: #f87171; font-size: 0.85rem; margin-top: 0.4rem; }
  `]
})
export class AppointmentRatingComponent implements OnInit {
  @Input() appointmentId!: string;
  @Input() status!: string;

  selectedRating = 0;
  comment = '';
  saving = false;
  submitted = false;
  error = '';
  existing: { rating: number; comment: string } | null = null;

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    if (this.status !== 'completed') return;
    this.http.get<{ rating: number; comment: string }>(ApiContract.appointments.rating(this.appointmentId)).subscribe({
      next: (r) => { this.existing = r; },
      error: () => {}
    });
  }

  submit(): void {
    if (!this.selectedRating) return;
    this.saving = true;
    this.error = '';
    this.http.post(ApiContract.appointments.rating(this.appointmentId), {
      rating: this.selectedRating,
      comment: this.comment
    }).subscribe({
      next: () => { this.submitted = true; this.saving = false; },
      error: (err) => { this.error = err?.error?.error ?? 'Could not submit rating'; this.saving = false; }
    });
  }
}

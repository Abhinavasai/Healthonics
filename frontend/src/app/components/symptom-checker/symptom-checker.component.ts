import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  AssistantService,
  SymptomCheckResponse,
  SymptomTrendEntry,
  SymptomTrendsResponse
} from '../../services/assistant.service';

@Component({
  selector: 'app-symptom-checker',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="symptom-checker">
      <div class="checker-header">
        <h2>AI Symptom Checker</h2>
        <p class="subtitle">Describe your symptoms and get an urgency assessment. This is not a substitute for professional medical advice.</p>
      </div>

      <!-- Worsening Alert -->
      <div *ngIf="trends?.worsening" class="worsening-alert">
        ⚠️ Your recent symptom checks show repeated urgent or emergency symptoms. Please book an appointment with your doctor.
      </div>

      <div class="checker-form">
        <div class="form-group">
          <label for="symptoms">Describe your symptoms</label>
          <textarea
            id="symptoms"
            [(ngModel)]="symptoms"
            placeholder="e.g. I have a severe headache, fever of 39°C, and stiff neck for the past 2 hours..."
            rows="5"
            maxlength="2000"
            [disabled]="loading"
          ></textarea>
          <div class="char-count">{{ symptoms.length }}/2000</div>
        </div>

        <div class="form-row">
          <div class="form-group half">
            <label for="age">Age (optional)</label>
            <input id="age" type="number" [(ngModel)]="age" placeholder="e.g. 35" min="0" max="120" [disabled]="loading" />
          </div>
          <div class="form-group half">
            <label for="gender">Biological sex (optional)</label>
            <select id="gender" [(ngModel)]="gender" [disabled]="loading">
              <option value="">Prefer not to say</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
            </select>
          </div>
        </div>

        <button class="btn-check" (click)="check()" [disabled]="loading || symptoms.trim().length < 10">
          <span *ngIf="!loading">Assess my symptoms</span>
          <span *ngIf="loading" class="spinner">Analysing...</span>
        </button>
      </div>

      <div *ngIf="error" class="error-banner">{{ error }}</div>

      <div *ngIf="result" class="result-card" [class]="urgencyClass(result.urgency ?? '')">
        <div class="urgency-badge">{{ urgencyLabel(result.urgency ?? '') }}</div>
        <div class="result-body">{{ result.reply }}</div>
        <div class="disclaimer"><strong>Disclaimer:</strong> {{ result.disclaimer }}</div>
        <div class="provider-info">Powered by {{ result.provider }}{{ result.fallback_used ? ' (rule-based fallback)' : '' }}</div>
        <button class="btn-reset" (click)="reset()">Check again</button>
      </div>

      <!-- Trend Timeline -->
      <div *ngIf="trends && trends.checks?.length" class="trend-section">
        <h3 class="trend-title">Your Symptom History</h3>
        <div class="trend-timeline">
          <div *ngFor="let entry of trends.checks" class="trend-item" [class]="'trend-' + entry.urgency">
            <div class="trend-dot"></div>
            <div class="trend-content">
              <div class="trend-meta">
                <span class="trend-urgency-badge">{{ urgencyLabel(entry.urgency) }}</span>
                <span class="trend-date">{{ entry.checked_at | date:'MMM d, y, h:mm a' }}</span>
              </div>
              <div class="trend-symptoms">{{ entry.symptoms | slice:0:120 }}{{ entry.symptoms.length > 120 ? '...' : '' }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .symptom-checker { max-width: 720px; margin: 32px auto; padding: 0 16px; }
    .checker-header h2 { font-size: 1.6rem; font-weight: 700; color: #1a202c; margin-bottom: 8px; }
    .subtitle { color: #718096; margin-bottom: 24px; }
    .worsening-alert {
      background: #fff5f5; border: 1px solid #fed7d7; color: #c53030;
      border-radius: 10px; padding: 12px 16px; margin-bottom: 16px; font-weight: 600;
    }
    .checker-form { background: #fff; border-radius: 12px; padding: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
    .form-group { margin-bottom: 16px; }
    .form-row { display: flex; gap: 16px; }
    .form-group.half { flex: 1; }
    label { display: block; font-weight: 600; font-size: 0.875rem; color: #4a5568; margin-bottom: 6px; }
    textarea, input, select {
      width: 100%; border: 1px solid #e2e8f0; border-radius: 8px;
      padding: 10px 12px; font-size: 0.95rem; outline: none;
      transition: border-color 0.2s; box-sizing: border-box;
    }
    textarea:focus, input:focus, select:focus { border-color: #4299e1; }
    textarea { resize: vertical; }
    .char-count { text-align: right; font-size: 0.75rem; color: #a0aec0; margin-top: 4px; }
    .btn-check {
      width: 100%; background: #4299e1; color: #fff; border: none; border-radius: 8px;
      padding: 12px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 8px; transition: background 0.2s;
    }
    .btn-check:hover:not(:disabled) { background: #3182ce; }
    .btn-check:disabled { background: #bee3f8; cursor: not-allowed; }
    .error-banner { margin-top: 16px; background: #fff5f5; border: 1px solid #fed7d7; color: #c53030; border-radius: 8px; padding: 12px 16px; }
    .result-card { margin-top: 24px; background: #fff; border-radius: 12px; padding: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); border-left: 4px solid #4299e1; }
    .result-card.emergency { border-left-color: #e53e3e; }
    .result-card.urgent { border-left-color: #ed8936; }
    .result-card.routine { border-left-color: #48bb78; }
    .result-card.selfcare { border-left-color: #68d391; }
    .urgency-badge { display: inline-block; font-weight: 700; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.05em; padding: 4px 10px; border-radius: 20px; background: #ebf8ff; color: #2b6cb0; margin-bottom: 16px; }
    .result-card.emergency .urgency-badge { background: #fff5f5; color: #c53030; }
    .result-card.urgent .urgency-badge { background: #fffaf0; color: #c05621; }
    .result-card.routine .urgency-badge { background: #f0fff4; color: #276749; }
    .result-body { line-height: 1.7; white-space: pre-wrap; margin-bottom: 16px; }
    .disclaimer { font-size: 0.8rem; color: #718096; background: #f7fafc; border-radius: 6px; padding: 10px 12px; margin-bottom: 12px; }
    .provider-info { font-size: 0.75rem; color: #a0aec0; margin-bottom: 16px; }
    .btn-reset { background: transparent; border: 1px solid #e2e8f0; border-radius: 8px; padding: 8px 16px; cursor: pointer; font-size: 0.875rem; color: #4a5568; }
    .btn-reset:hover { background: #f7fafc; }
    .trend-section { margin-top: 40px; }
    .trend-title { font-size: 1.1rem; font-weight: 700; color: #2d3748; margin-bottom: 16px; }
    .trend-timeline { display: flex; flex-direction: column; gap: 0; }
    .trend-item { display: flex; gap: 16px; align-items: flex-start; position: relative; padding-bottom: 20px; }
    .trend-item:not(:last-child)::before {
      content: ''; position: absolute; left: 7px; top: 16px;
      width: 2px; bottom: 0; background: #e2e8f0;
    }
    .trend-dot { width: 16px; height: 16px; border-radius: 50%; flex-shrink: 0; margin-top: 2px; background: #cbd5e0; }
    .trend-emergency .trend-dot { background: #fc8181; }
    .trend-urgent .trend-dot { background: #f6ad55; }
    .trend-routine .trend-dot { background: #68d391; }
    .trend-selfcare .trend-dot { background: #9ae6b4; }
    .trend-content { flex: 1; background: #fff; border-radius: 10px; padding: 12px 14px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
    .trend-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; flex-wrap: wrap; }
    .trend-urgency-badge { font-size: 0.7rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; padding: 2px 8px; border-radius: 20px; background: #edf2f7; color: #4a5568; }
    .trend-emergency .trend-urgency-badge { background: #fff5f5; color: #c53030; }
    .trend-urgent .trend-urgency-badge { background: #fffaf0; color: #c05621; }
    .trend-routine .trend-urgency-badge { background: #f0fff4; color: #276749; }
    .trend-date { font-size: 0.75rem; color: #a0aec0; }
    .trend-symptoms { font-size: 0.875rem; color: #4a5568; line-height: 1.5; }
    .spinner { opacity: 0.7; }
  `]
})
export class SymptomCheckerComponent implements OnInit {
  symptoms = '';
  age: number | undefined;
  gender = '';
  loading = false;
  error = '';
  result: (SymptomCheckResponse & { urgency?: string }) | null = null;
  trends: SymptomTrendsResponse | null = null;

  constructor(private assistant: AssistantService) {}

  ngOnInit(): void {
    this.loadTrends();
  }

  loadTrends(): void {
    this.assistant.symptomTrends().subscribe({
      next: (t) => { this.trends = t; },
      error: () => {}
    });
  }

  urgencyClass(urgency: string): string {
    const map: Record<string, string> = { emergency: 'emergency', urgent: 'urgent', selfcare: 'selfcare', routine: 'routine' };
    return map[urgency] ?? 'routine';
  }

  urgencyLabel(urgency: string): string {
    const map: Record<string, string> = { emergency: 'Emergency', urgent: 'Urgent', selfcare: 'Self-care', routine: 'Routine' };
    return map[urgency] ?? 'Assessment';
  }

  check(): void {
    if (this.symptoms.trim().length < 10) return;
    this.loading = true;
    this.error = '';
    this.result = null;
    this.assistant
      .symptomCheck(this.symptoms.trim(), this.age, this.gender || undefined)
      .subscribe({
        next: (res) => {
          this.result = res;
          this.loading = false;
          this.loadTrends();
        },
        error: (err) => {
          this.error = err?.error?.error ?? 'Failed to analyse symptoms. Please try again.';
          this.loading = false;
        }
      });
  }

  reset(): void {
    this.result = null;
    this.symptoms = '';
    this.age = undefined;
    this.gender = '';
    this.error = '';
  }
}

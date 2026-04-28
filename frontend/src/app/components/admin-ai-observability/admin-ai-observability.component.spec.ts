import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { AdminAiObservabilityComponent } from './admin-ai-observability.component';
import { AdminAiRuntimeService } from '../../services/admin-ai-runtime.service';

describe('AdminAiObservabilityComponent', () => {
  let fixture: ComponentFixture<AdminAiObservabilityComponent>;
  let serviceSpy: jasmine.SpyObj<AdminAiRuntimeService>;

  beforeEach(async () => {
    serviceSpy = jasmine.createSpyObj<AdminAiRuntimeService>('AdminAiRuntimeService', [
      'getObservability',
      'updateSettings',
      'runEval'
    ]);
    serviceSpy.getObservability.and.returnValue(
      of({
        settings: {
          ollama_enabled: false,
          fallback_enabled: true,
          rate_limit_enabled: true,
          rate_limit_per_minute: 30,
          cache_enabled: true,
          cache_ttl_seconds: 900,
          ollama_model: 'llama3.1:8b',
          updated_at: '2026-04-28T00:00:00Z',
          updated_by: ''
        },
        observability: {
          queued_jobs_24h: 2,
          completed_jobs_24h: 2,
          failed_jobs_24h: 0,
          avg_latency_seconds: 0.25,
          cache_ready_count: 4,
          pending_docs_count: 1,
          failed_docs_count: 0
        },
        runtime: {
          ai_enabled: true,
          ollama_host: 'http://127.0.0.1:11434',
          configured_model: 'llama3.1:8b',
          ollama_reachable: true,
          model_available: true,
          active_mode: 'ollama'
        }
      })
    );

    await TestBed.configureTestingModule({
      imports: [AdminAiObservabilityComponent],
      providers: [{ provide: AdminAiRuntimeService, useValue: serviceSpy }]
    }).compileComponents();

    fixture = TestBed.createComponent(AdminAiObservabilityComponent);
    fixture.detectChanges();
  });

  it('loads and renders page', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(serviceSpy.getObservability).toHaveBeenCalled();
    expect(el.querySelector('[data-cy="admin-ai-observability-page"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="ai-runtime-health"]')).toBeTruthy();
  });

  it('saves settings', () => {
    serviceSpy.updateSettings.and.returnValue(
      of({
        settings: {
          ollama_enabled: true,
          fallback_enabled: true,
          rate_limit_enabled: true,
          rate_limit_per_minute: 20,
          cache_enabled: true,
          cache_ttl_seconds: 600,
          ollama_model: 'llama3.1:8b',
          updated_at: '2026-04-28T00:00:00Z',
          updated_by: 'admin'
        },
        observability: {
          queued_jobs_24h: 1,
          completed_jobs_24h: 1,
          failed_jobs_24h: 0,
          avg_latency_seconds: 0.1,
          cache_ready_count: 5,
          pending_docs_count: 0,
          failed_docs_count: 0
        }
      })
    );

    const component = fixture.componentInstance;
    component.form.ollama_enabled = true;
    component.form.rate_limit_per_minute = 20;
    component.save();

    expect(serviceSpy.updateSettings).toHaveBeenCalled();
  });

  it('runs eval', () => {
    serviceSpy.runEval.and.returnValue(
      of({
        eval: {
          model_name: 'llama3.1:8b',
          samples_evaluated: 8,
          success_rate: 0.75,
          quality_score: 75
        }
      })
    );
    const component = fixture.componentInstance;
    component.runEval();
    expect(serviceSpy.runEval).toHaveBeenCalled();
    expect(component.evalResult?.samples_evaluated).toBe(8);
  });
});

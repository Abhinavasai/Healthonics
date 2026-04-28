import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { AdminAiRuntimeService } from './admin-ai-runtime.service';
import { ApiContract } from './api-contract';

describe('AdminAiRuntimeService', () => {
  let service: AdminAiRuntimeService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [AdminAiRuntimeService, provideHttpClient(), provideHttpClientTesting()]
    });
    service = TestBed.inject(AdminAiRuntimeService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('loads observability payload', () => {
    service.getObservability().subscribe((res) => {
      expect(res.settings.ollama_model).toBe('llama3.1:8b');
      expect(res.observability.queued_jobs_24h).toBe(4);
    });

    const req = httpMock.expectOne(ApiContract.admin.aiObservability);
    expect(req.request.method).toBe('GET');
    req.flush({
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
        queued_jobs_24h: 4,
        completed_jobs_24h: 3,
        failed_jobs_24h: 1,
        avg_latency_seconds: 0.45,
        cache_ready_count: 9,
        pending_docs_count: 1,
        failed_docs_count: 0
      }
    });
  });

  it('updates runtime settings', () => {
    service
      .updateSettings({
        ollama_enabled: true,
        fallback_enabled: true,
        rate_limit_enabled: true,
        rate_limit_per_minute: 20,
        cache_enabled: true,
        cache_ttl_seconds: 1200,
        ollama_model: 'llama3.1:8b'
      })
      .subscribe((res) => {
        expect(res.settings.ollama_enabled).toBeTrue();
        expect(res.settings.rate_limit_per_minute).toBe(20);
      });

    const req = httpMock.expectOne(ApiContract.admin.aiSettings);
    expect(req.request.method).toBe('PUT');
    expect(req.request.body.rate_limit_per_minute).toBe(20);
    req.flush({
      settings: {
        ollama_enabled: true,
        fallback_enabled: true,
        rate_limit_enabled: true,
        rate_limit_per_minute: 20,
        cache_enabled: true,
        cache_ttl_seconds: 1200,
        ollama_model: 'llama3.1:8b',
        updated_at: '2026-04-28T00:00:00Z',
        updated_by: 'admin-id'
      },
      observability: {
        queued_jobs_24h: 5,
        completed_jobs_24h: 4,
        failed_jobs_24h: 1,
        avg_latency_seconds: 0.4,
        cache_ready_count: 10,
        pending_docs_count: 0,
        failed_docs_count: 0
      }
    });
  });

  it('runs eval', () => {
    service.runEval().subscribe((res) => {
      expect(res.eval.model_name).toBe('llama3.1:8b');
      expect(res.eval.samples_evaluated).toBe(12);
    });

    const req = httpMock.expectOne(ApiContract.admin.aiEval);
    expect(req.request.method).toBe('POST');
    req.flush({
      eval: {
        model_name: 'llama3.1:8b',
        samples_evaluated: 12,
        success_rate: 0.75,
        quality_score: 75
      }
    });
  });
});

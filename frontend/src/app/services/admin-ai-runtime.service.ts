import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export type AdminAiRuntimeSettings = {
  ollama_enabled: boolean;
  fallback_enabled: boolean;
  rate_limit_enabled: boolean;
  rate_limit_per_minute: number;
  cache_enabled: boolean;
  cache_ttl_seconds: number;
  ollama_model: string;
  updated_at: string;
  updated_by: string;
};

export type AdminAiObservability = {
  queued_jobs_24h: number;
  completed_jobs_24h: number;
  failed_jobs_24h: number;
  avg_latency_seconds: number;
  cache_ready_count: number;
  pending_docs_count: number;
  failed_docs_count: number;
};

export type AdminAiObservabilityResponse = {
  settings: AdminAiRuntimeSettings;
  observability: AdminAiObservability;
  runtime?: {
    ai_enabled: boolean;
    ollama_host: string;
    configured_model: string;
    ollama_reachable: boolean;
    model_available: boolean;
    active_mode: string;
  };
};

export type AdminAiEvalResponse = {
  eval: {
    model_name: string;
    samples_evaluated: number;
    success_rate: number;
    quality_score: number;
  };
  runtime?: {
    ai_enabled: boolean;
    ollama_host: string;
    configured_model: string;
    ollama_reachable: boolean;
    model_available: boolean;
    active_mode: string;
  };
};

@Injectable({ providedIn: 'root' })
export class AdminAiRuntimeService {
  constructor(private readonly http: HttpClient) {}

  getObservability(): Observable<AdminAiObservabilityResponse> {
    return this.http.get<AdminAiObservabilityResponse>(ApiContract.admin.aiObservability);
  }

  updateSettings(payload: {
    ollama_enabled: boolean;
    fallback_enabled: boolean;
    rate_limit_enabled: boolean;
    rate_limit_per_minute: number;
    cache_enabled: boolean;
    cache_ttl_seconds: number;
    ollama_model: string;
  }): Observable<AdminAiObservabilityResponse> {
    return this.http.put<AdminAiObservabilityResponse>(ApiContract.admin.aiSettings, payload);
  }

  runEval(): Observable<AdminAiEvalResponse> {
    return this.http.post<AdminAiEvalResponse>(ApiContract.admin.aiEval, {});
  }
}

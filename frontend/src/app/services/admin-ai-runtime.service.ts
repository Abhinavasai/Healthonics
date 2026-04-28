import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { AdminAiEvalResponse, AdminAiObservabilityResponse } from './api-types';
export type {
  AdminAiRuntimeSettings,
  AdminAiObservability,
  AdminAiRuntimeStatus,
  AdminAiObservabilityResponse,
  AdminAiEvalResponse
} from './api-types';

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

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type {
  KnowledgeDetail,
  KnowledgeDocumentsListResponse,
  KnowledgeVersionsResponse,
  SimilarityScanResponse
} from './api-types';
export type {
  KnowledgeListDocument,
  KnowledgeDocumentsListResponse,
  KnowledgeDetail,
  KnowledgeVersionRow,
  KnowledgeVersionsResponse,
  SimilarityPair,
  SimilarityScanResponse
} from './api-types';

@Injectable({ providedIn: 'root' })
export class KnowledgeAdminService {
  constructor(private http: HttpClient) {}

  listDocs(): Observable<KnowledgeDocumentsListResponse> {
    return this.http.get<KnowledgeDocumentsListResponse>(ApiContract.admin.knowledgeDocs);
  }

  getDoc(id: string): Observable<KnowledgeDetail> {
    return this.http.get<KnowledgeDetail>(ApiContract.admin.knowledgeDocById(id));
  }

  createDoc(payload: {
    title: string;
    body: string;
    review_interval_days?: number;
  }): Observable<{ id: string }> {
    return this.http.post<{ id: string }>(ApiContract.admin.knowledgeDocs, payload);
  }

  markReviewed(id: string): Observable<{ ok: boolean }> {
    return this.http.post<{ ok: boolean }>(ApiContract.admin.knowledgeDocReview(id), {});
  }

  patchDoc(
    id: string,
    body: Partial<{ title: string; body: string; review_interval_days: number }>
  ): Observable<{ id: string; current_version: number }> {
    return this.http.patch<{ id: string; current_version: number }>(
      ApiContract.admin.knowledgeDocById(id),
      body
    );
  }

  listVersions(id: string): Observable<KnowledgeVersionsResponse> {
    return this.http.get<KnowledgeVersionsResponse>(ApiContract.admin.knowledgeDocVersions(id));
  }

  similarityScan(threshold?: number, maxChunks?: number): Observable<SimilarityScanResponse> {
    const params: string[] = [];
    if (threshold != null) {
      params.push(`threshold=${encodeURIComponent(String(threshold))}`);
    }
    if (maxChunks != null) {
      params.push(`max_chunks=${encodeURIComponent(String(maxChunks))}`);
    }
    const q = params.length ? `?${params.join('&')}` : '';
    return this.http.get<SimilarityScanResponse>(`${ApiContract.admin.knowledgeSimilarityScan}${q}`);
  }
}

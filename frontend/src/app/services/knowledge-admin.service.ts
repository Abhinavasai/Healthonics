import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export interface KnowledgeListDocument {
  id: string;
  title: string;
  excerpt: string;
  created_at: string;
  updated_at: string;
  current_version?: number;
  review_interval_days?: number;
  last_reviewed_at?: string;
  days_since_review?: number;
  health_score?: number;
  is_stale?: boolean;
}

export interface KnowledgeDocumentsListResponse {
  documents: KnowledgeListDocument[];
}

/** GET /admin/knowledge-docs/:id (PR-28) */
export interface KnowledgeDetail {
  id: string;
  title: string;
  body: string;
  updated_at: string;
  current_version: number;
  review_interval_days: number;
  last_reviewed_at?: string;
  days_since_review: number;
  health_score: number;
  is_stale: boolean;
}

export interface KnowledgeVersionRow {
  version: number;
  title: string;
  excerpt: string;
  created_at: string;
  created_by: string;
}

export interface KnowledgeVersionsResponse {
  versions: KnowledgeVersionRow[];
}

export interface SimilarityPair {
  similarity: number;
  document_a_id: string;
  document_b_id: string;
  title_a: string;
  title_b: string;
  chunk_a_index: number;
  chunk_b_index: number;
  excerpt_a: string;
  excerpt_b: string;
  conflict_hint: string;
}

export interface SimilarityScanResponse {
  threshold: number;
  chunks_scanned: number;
  pairs: SimilarityPair[];
}

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

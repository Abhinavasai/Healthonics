import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, catchError, map, of } from 'rxjs';
import { ApiContract } from './api-contract';

/**
 * Doctor view of patient-uploaded documents + AI summary (Sprint 3 feature 4).
 * Backend (Rohith): GET/POST /api/doctor/documents...
 */
export type SummaryStatus = 'none' | 'pending' | 'ready' | 'failed';

export interface DoctorDocumentListItem {
  id: string;
  filename: string;
  patient_email: string;
  created_at: string;
  summary_status?: SummaryStatus;
}

export interface DoctorDocumentDetail {
  id: string;
  filename: string;
  patient_id: string;
  patient_email: string;
  size_bytes: number;
  content_type: string;
  created_at: string;
  summary: string | null;
  summary_status: SummaryStatus;
  /** Present when summary_status is failed */
  summary_error?: string | null;
}

export interface DoctorDocumentsListResponse {
  documents: DoctorDocumentListItem[];
}

@Injectable({ providedIn: 'root' })
export class DoctorDocumentsService {
  constructor(private http: HttpClient) {}

  list(): Observable<DoctorDocumentListItem[]> {
    return this.http.get<DoctorDocumentsListResponse>(ApiContract.doctorDocuments.base).pipe(
      map((r) => r.documents ?? []),
      catchError(() => of([]))
    );
  }

  getDetail(documentId: string): Observable<DoctorDocumentDetail | null> {
    return this.http
      .get<DoctorDocumentDetail>(ApiContract.doctorDocuments.byId(encodeURIComponent(documentId)))
      .pipe(
      catchError(() => of(null))
    );
  }

  /** Triggers async summarization; poll getDetail until ready/failed. */
  requestSummary(documentId: string): Observable<{ status?: string } | null> {
    return this.http
      .post<{ status?: string }>(
        ApiContract.doctorDocuments.summarize(encodeURIComponent(documentId)),
        {}
      )
      .pipe(catchError(() => of(null)));
  }
}

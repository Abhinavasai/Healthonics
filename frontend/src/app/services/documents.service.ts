import { Injectable } from '@angular/core';
import { HttpClient, HttpEvent, HttpEventType } from '@angular/common/http';
import { Observable, catchError, filter, map, of } from 'rxjs';

/**
 * Patient medical documents — aligns with Sprint 3 backend Kaushik:
 * GET/POST /api/documents, GET /api/documents/:id/download
 */
export interface PatientDocument {
  id: string;
  filename: string;
  content_type?: string;
  size_bytes?: number;
  status?: 'pending' | 'ready' | 'failed';
  created_at: string;
}

export interface DocumentsListResponse {
  documents: PatientDocument[];
}

@Injectable({ providedIn: 'root' })
export class DocumentsService {
  private readonly API = '/api/documents';

  constructor(private http: HttpClient) {}

  list(): Observable<PatientDocument[]> {
    return this.http.get<DocumentsListResponse>(this.API).pipe(
      map((r) => r.documents ?? []),
      catchError(() => of([]))
    );
  }

  /**
   * Multipart upload; backend expects field name `file`.
   * Emits upload progress 0–100 then the created document on completion.
   */
  upload(
    file: File
  ): Observable<{ type: 'progress'; percent: number } | { type: 'done'; doc: PatientDocument } | { type: 'error'; message: string }> {
    const fd = new FormData();
    fd.append('file', file, file.name);

    return this.http
      .post<PatientDocument>(this.API, fd, {
        reportProgress: true,
        observe: 'events'
      })
      .pipe(
        map((event: HttpEvent<PatientDocument>) => {
          if (event.type === HttpEventType.UploadProgress && event.total) {
            const percent = Math.round((100 * event.loaded) / event.total);
            return { type: 'progress' as const, percent };
          }
          if (event.type === HttpEventType.Response && event.body) {
            return { type: 'done' as const, doc: event.body };
          }
          return null;
        }),
        filter((x): x is NonNullable<typeof x> => x !== null),
        catchError((err) => {
          const msg = err?.error?.error ?? err?.message ?? 'Upload failed';
          return of({ type: 'error' as const, message: String(msg) });
        })
      );
  }

  /** GET file stream; opens in new tab via blob URL in the component. */
  downloadBlob(id: string): Observable<Blob> {
    return this.http.get(`${this.API}/${encodeURIComponent(id)}/download`, {
      responseType: 'blob'
    });
  }
}

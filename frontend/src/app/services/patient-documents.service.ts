import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PatientDocumentRow {
  id: string;
  filename: string;
  size_bytes: number;
  content_type: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class PatientDocumentsService {
  private readonly api = '/api/documents';

  constructor(private http: HttpClient) {}

  list(): Observable<{ documents: PatientDocumentRow[] }> {
    return this.http.get<{ documents: PatientDocumentRow[] }>(this.api);
  }

  upload(file: File): Observable<PatientDocumentRow> {
    const formData = new FormData();
    formData.append('file', file);
    return this.http.post<PatientDocumentRow>(this.api, formData);
  }

  downloadUrl(id: string): string {
    return `${this.api}/${encodeURIComponent(id)}/download`;
  }
}


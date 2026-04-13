import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PatientUploadResponse {
  id: string;
  filename: string;
}

@Injectable({ providedIn: 'root' })
export class PatientDocumentsService {
  private readonly API = '/api';

  constructor(private http: HttpClient) {}

  /** Multipart upload; field name must be `file` (matches backend). */
  upload(file: File): Observable<PatientUploadResponse> {
    const body = new FormData();
    body.append('file', file, file.name);
    return this.http.post<PatientUploadResponse>(`${this.API}/patient/documents`, body);
  }
}

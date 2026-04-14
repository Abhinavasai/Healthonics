import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PatientFileRow {
  id: string;
  patient_id: string;
  uploaded_by: string;
  description: string;
  original_name: string;
  mime_type: string;
  byte_size: number;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class PatientFilesService {
  private readonly API = '/api';

  constructor(private http: HttpClient) {}

  list(patientId: string): Observable<{ files: PatientFileRow[] }> {
    return this.http.get<{ files: PatientFileRow[] }>(`${this.API}/patients/${patientId}/files`);
  }

  upload(patientId: string, file: File, description: string): Observable<unknown> {
    const fd = new FormData();
    fd.append('file', file);
    fd.append('description', description);
    return this.http.post(`${this.API}/patients/${patientId}/files`, fd);
  }

  download(fileId: string): Observable<Blob> {
    return this.http.get(`${this.API}/files/${fileId}`, { responseType: 'blob' });
  }
}

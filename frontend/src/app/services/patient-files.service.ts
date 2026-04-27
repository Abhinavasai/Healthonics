import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

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
  constructor(private http: HttpClient) {}

  list(patientId: string): Observable<{ files: PatientFileRow[] }> {
    return this.http.get<{ files: PatientFileRow[] }>(ApiContract.patientFiles.list(patientId));
  }

  upload(patientId: string, file: File, description: string): Observable<unknown> {
    const fd = new FormData();
    fd.append('file', file);
    fd.append('description', description);
    return this.http.post(ApiContract.patientFiles.upload(patientId), fd);
  }

  download(fileId: string): Observable<Blob> {
    return this.http.get(ApiContract.patientFiles.download(fileId), { responseType: 'blob' });
  }
}

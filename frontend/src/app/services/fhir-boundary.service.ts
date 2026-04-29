import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { AdminFhirExportResponse, AdminFhirImportResponse, FhirBundle } from './api-types';

@Injectable({ providedIn: 'root' })
export class FhirBoundaryService {
  constructor(private readonly http: HttpClient) {}

  exportBundle(): Observable<AdminFhirExportResponse> {
    return this.http.get<AdminFhirExportResponse>(ApiContract.admin.fhirExport);
  }

  importBundle(bundle: FhirBundle, apply = false): Observable<AdminFhirImportResponse> {
    return this.http.post<AdminFhirImportResponse>(ApiContract.admin.fhirImport, { bundle, apply });
  }
}


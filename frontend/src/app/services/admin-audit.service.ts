import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { ApiListResponse, AuditLogEntry } from './api-types';
export type { AuditLogEntry } from './api-types';

@Injectable({ providedIn: 'root' })
export class AdminAuditService {
  constructor(private http: HttpClient) {}

  list(): Observable<ApiListResponse<AuditLogEntry, 'entries'>> {
    return this.http.get<ApiListResponse<AuditLogEntry, 'entries'>>(ApiContract.admin.auditLog);
  }
}

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export interface AuditLogEntry {
  id: string;
  actor_user_id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  detail: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class AdminAuditService {
  constructor(private http: HttpClient) {}

  list(): Observable<{ entries: AuditLogEntry[] }> {
    return this.http.get<{ entries: AuditLogEntry[] }>(ApiContract.admin.auditLog);
  }
}

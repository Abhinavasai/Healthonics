import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { AdminLifecycleSettingsKpisResponse } from './api-types';
export type { AdminLifecycleSettings, AdminLifecycleKpis, AdminLifecycleSettingsKpisResponse } from './api-types';

@Injectable({ providedIn: 'root' })
export class AdminUserLifecycleService {
  constructor(private http: HttpClient) {}

  getSettingsKpis(): Observable<AdminLifecycleSettingsKpisResponse> {
    return this.http.get<AdminLifecycleSettingsKpisResponse>(ApiContract.admin.userLifecycleSettingsKpis);
  }

  updateSettings(payload: {
    new_user_window_days: number;
    inactive_window_days: number;
  }): Observable<AdminLifecycleSettingsKpisResponse> {
    return this.http.put<AdminLifecycleSettingsKpisResponse>(ApiContract.admin.userLifecycleSettings, payload);
  }
}

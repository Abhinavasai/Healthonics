import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export type AdminLifecycleSettings = {
  new_user_window_days: number;
  inactive_window_days: number;
  updated_at: string;
  updated_by: string;
};

export type AdminLifecycleKpis = {
  total_users: number;
  total_patients: number;
  total_doctors: number;
  total_admins: number;
  new_users_in_window: number;
  inactive_users_count: number;
};

export type AdminLifecycleSettingsKpisResponse = {
  settings: AdminLifecycleSettings;
  kpis: AdminLifecycleKpis;
};

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

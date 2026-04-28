import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { AdminLifecycleSettingsKpisResponse } from './api-types';
export type { AdminLifecycleSettings, AdminLifecycleKpis, AdminLifecycleSettingsKpisResponse } from './api-types';

export type AdminLifecycleUser = {
  id: string;
  email: string;
  role: 'patient' | 'doctor' | 'admin';
  is_active: boolean;
  created_at: string;
  deactivated_at: string;
  deactivated_by: string;
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

  listUsers(): Observable<{ users: AdminLifecycleUser[] }> {
    return this.http.get<{ users: AdminLifecycleUser[] }>(ApiContract.admin.userLifecycleUsers);
  }

  createUser(payload: {
    email: string;
    password: string;
    role: 'patient' | 'doctor' | 'admin';
  }): Observable<{ id: string; email: string; role: string; is_active: boolean }> {
    return this.http.post<{ id: string; email: string; role: string; is_active: boolean }>(
      ApiContract.admin.userLifecycleUsers,
      payload
    );
  }

  updateUser(
    id: string,
    payload: { email?: string; role?: 'patient' | 'doctor' | 'admin' }
  ): Observable<{ id: string; email: string; role: string; is_active: boolean }> {
    return this.http.put<{ id: string; email: string; role: string; is_active: boolean }>(
      ApiContract.admin.userLifecycleUserById(id),
      payload
    );
  }

  deactivateUser(id: string): Observable<{ id: string; status: string }> {
    return this.http.patch<{ id: string; status: string }>(ApiContract.admin.userLifecycleDeactivate(id), {});
  }

  resetPassword(id: string, newPassword: string): Observable<{ id: string; password_reset: boolean }> {
    return this.http.post<{ id: string; password_reset: boolean }>(
      ApiContract.admin.userLifecycleResetPassword(id),
      { new_password: newPassword }
    );
  }
}

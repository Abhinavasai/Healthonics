import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PatientDashboard {
  upcoming_count: number;
  pending_count: number;
  unread_messages: number;
  next_appointment?: {
    id: string;
    scheduled_at: string;
    status: string;
    doctor_email: string;
  };
}

export interface DoctorDashboard {
  today_appointments_count: number;
  pending_queue_count: number;
  unread_messages: number;
}

export interface AdminDashboard {
  db_ok: boolean;
  patients_count: number;
  doctors_count: number;
  admins_count: number;
}

export interface AdminUserRow {
  id: string;
  email: string;
  role: string;
  active: boolean;
}

export interface NotificationPreferences {
  email_appointment_reminders: boolean;
  sms_appointment_reminders: boolean;
}

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private readonly api = '/api';

  constructor(private http: HttpClient) {}

  patient(): Observable<PatientDashboard> {
    return this.http.get<PatientDashboard>(`${this.api}/dashboard/patient`);
  }

  doctor(): Observable<DoctorDashboard> {
    return this.http.get<DoctorDashboard>(`${this.api}/dashboard/doctor`);
  }

  admin(): Observable<AdminDashboard> {
    return this.http.get<AdminDashboard>(`${this.api}/dashboard/admin`);
  }

  listUsers(): Observable<{ users: AdminUserRow[] }> {
    return this.http.get<{ users: AdminUserRow[] }>(`${this.api}/admin/users`);
  }

  setUserActive(userId: string, active: boolean): Observable<{ ok: boolean }> {
    return this.http.patch<{ ok: boolean }>(`${this.api}/admin/users/${userId}`, { active });
  }

  getNotificationPreferences(): Observable<NotificationPreferences> {
    return this.http.get<NotificationPreferences>(`${this.api}/me/notification-preferences`);
  }

  putNotificationPreferences(p: NotificationPreferences): Observable<NotificationPreferences> {
    return this.http.put<NotificationPreferences>(`${this.api}/me/notification-preferences`, p);
  }
}

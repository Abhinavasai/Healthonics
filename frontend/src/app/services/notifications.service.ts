import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';

export interface NotificationRow {
  id: string;
  title: string;
  body: string;
  channel: string;
  status: string;
  scheduled_for?: string | null;
  created_at: string;
}

export interface NotificationPreferenceRow {
  category: string;
  enabled: boolean;
  email_enabled: boolean;
  sms_enabled: boolean;
  in_app_enabled: boolean;
}

@Injectable({ providedIn: 'root' })
export class NotificationsInboxService {
  constructor(private http: HttpClient) {}

  listMine(): Observable<{ notifications: NotificationRow[] }> {
    return this.http.get<{ notifications: NotificationRow[] }>(ApiContract.notifications.listMine);
  }

  listPreferences(): Observable<{ preferences: NotificationPreferenceRow[] }> {
    return this.http.get<{ preferences: NotificationPreferenceRow[] }>(
      ApiContract.notifications.listPreferences
    );
  }

  savePreferences(preferences: NotificationPreferenceRow[]): Observable<{ updated: number }> {
    return this.http.put<{ updated: number }>(ApiContract.notifications.upsertPreferences, {
      preferences
    });
  }
}

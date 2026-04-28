import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type { ApiListResponse, NotificationPreferenceRow, NotificationRow } from './api-types';
export type { NotificationRow, NotificationPreferenceRow } from './api-types';

@Injectable({ providedIn: 'root' })
export class NotificationsInboxService {
  constructor(private http: HttpClient) {}

  listMine(): Observable<ApiListResponse<NotificationRow, 'notifications'>> {
    return this.http.get<ApiListResponse<NotificationRow, 'notifications'>>(
      ApiContract.notifications.listMine
    );
  }

  listPreferences(): Observable<ApiListResponse<NotificationPreferenceRow, 'preferences'>> {
    return this.http.get<ApiListResponse<NotificationPreferenceRow, 'preferences'>>(
      ApiContract.notifications.listPreferences
    );
  }

  savePreferences(preferences: NotificationPreferenceRow[]): Observable<{ updated: number }> {
    return this.http.put<{ updated: number }>(ApiContract.notifications.upsertPreferences, {
      preferences
    });
  }
}

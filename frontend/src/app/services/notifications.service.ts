import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApiContract } from './api-contract';
import type {
  AdminNotificationRow,
  AdminNotificationsSummary,
  ApiListResponse,
  NotificationConsentEventRow,
  NotificationPreferenceRow,
  NotificationRow
} from './api-types';
export type {
  NotificationRow,
  NotificationPreferenceRow,
  AdminNotificationRow,
  AdminNotificationsSummary,
  NotificationConsentEventRow
} from './api-types';

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

  listAdminNotifications(): Observable<ApiListResponse<AdminNotificationRow, 'notifications'>> {
    return this.http.get<ApiListResponse<AdminNotificationRow, 'notifications'>>(
      ApiContract.admin.notifications
    );
  }

  adminSummary(): Observable<AdminNotificationsSummary> {
    return this.http.get<AdminNotificationsSummary>(ApiContract.admin.notificationsSummary);
  }

  listAdminConsentHistory(): Observable<ApiListResponse<NotificationConsentEventRow, 'events'>> {
    return this.http.get<ApiListResponse<NotificationConsentEventRow, 'events'>>(
      ApiContract.admin.notificationsConsentHistory
    );
  }

  retryFailedNotification(id: string): Observable<{ id: string; status: string }> {
    return this.http.post<{ id: string; status: string }>(ApiContract.admin.notificationRetry(id), {});
  }
}

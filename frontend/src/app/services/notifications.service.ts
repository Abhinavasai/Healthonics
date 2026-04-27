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

@Injectable({ providedIn: 'root' })
export class NotificationsInboxService {
  constructor(private http: HttpClient) {}

  listMine(): Observable<{ notifications: NotificationRow[] }> {
    return this.http.get<{ notifications: NotificationRow[] }>(ApiContract.notifications.listMine);
  }
}

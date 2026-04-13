import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

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
  private readonly API = '/api';

  constructor(private http: HttpClient) {}

  listMine(): Observable<{ notifications: NotificationRow[] }> {
    return this.http.get<{ notifications: NotificationRow[] }>(`${this.API}/notifications`);
  }
}

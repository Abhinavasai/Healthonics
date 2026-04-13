import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface BootstrapPayload {
  app_name: string;
  api_version: string;
  shell_mobile_breakpoint_px?: number;
}

@Injectable({ providedIn: 'root' })
export class BootstrapService {
  constructor(private http: HttpClient) {}

  get(): Observable<BootstrapPayload> {
    return this.http.get<BootstrapPayload>('/api/bootstrap');
  }
}


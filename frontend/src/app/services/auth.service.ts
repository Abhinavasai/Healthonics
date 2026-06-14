import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';
import { ApiContract } from './api-contract';

export interface UserProfile {
  id: string;
  email: string;
  role: string;
}

export interface AuthResponse {
  token: string;
  refresh_token?: string;
  user: UserProfile;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  constructor(
    private http: HttpClient,
    private router: Router
  ) {}

  getToken(): string | null {
    return sessionStorage.getItem('token');
  }

  getRefreshToken(): string | null {
    return sessionStorage.getItem('refresh_token');
  }

  setAccessToken(token: string): void {
    sessionStorage.setItem('token', token);
  }

  getUser(): UserProfile | null {
    const raw = sessionStorage.getItem('user');
    if (!raw) return null;
    try {
      return JSON.parse(raw) as UserProfile;
    } catch {
      sessionStorage.clear();
      return null;
    }
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }

  register(email: string, password: string, role: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(ApiContract.auth.register, { email, password, role })
      .pipe(tap((res) => this.setSession(res)));
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(ApiContract.auth.login, { email, password })
      .pipe(tap((res) => this.setSession(res)));
  }

  logout(): void {
    const rt = this.getRefreshToken();
    if (rt) {
      this.http.post(ApiContract.auth.logout, { refresh_token: rt }).subscribe({ error: () => {} });
    }
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('refresh_token');
    sessionStorage.removeItem('user');
    this.router.navigate(['/login']);
  }

  refreshAccessToken(): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(ApiContract.auth.refresh, { refresh_token: this.getRefreshToken() })
      .pipe(tap((res) => this.setSession(res)));
  }

  private setSession(res: AuthResponse): void {
    sessionStorage.setItem('token', res.token);
    sessionStorage.setItem('user', JSON.stringify(res.user));
    if (res.refresh_token) {
      sessionStorage.setItem('refresh_token', res.refresh_token);
    }
  }

  redirectToDashboard(role: string): void {
    const path =
      role === 'admin'
        ? '/admin'
        : role === 'doctor'
          ? '/doctor/dashboard'
          : '/patient/dashboard';
    this.router.navigate([path]);
  }
}

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';

export interface UserProfile {
  id: string;
  email: string;
  role: string;
}

export interface AuthResponse {
  token: string;
  user: UserProfile;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly API = '/api';

  constructor(
    private http: HttpClient,
    private router: Router
  ) {}

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  getUser(): UserProfile | null {
    const raw = localStorage.getItem('user');
    return raw ? JSON.parse(raw) : null;
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }

  register(email: string, password: string, role: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(`${this.API}/register`, { email, password, role })
      .pipe(tap((res) => this.setSession(res)));
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return this.http
      .post<AuthResponse>(`${this.API}/login`, { email, password })
      .pipe(tap((res) => this.setSession(res)));
  }

  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    this.router.navigate(['/login']);
  }

  private setSession(res: AuthResponse): void {
    localStorage.setItem('token', res.token);
    localStorage.setItem('user', JSON.stringify(res.user));
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

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss'
})
export class LoginComponent implements OnInit {
  form: FormGroup;
  error = '';
  loading = false;
  showPassword = false;
  emailFocused = false;
  passwordFocused = false;

  constructor(
    private fb: FormBuilder,
    private auth: AuthService
  ) {
    this.form = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required]]
    });
  }

  ngOnInit(): void {
    if (this.auth.isLoggedIn()) {
      const user = this.auth.getUser();
      const role = user && typeof (user as any).role === 'string' && (user as any).role.trim() !== '' ? (user as any).role : null;
      if (role) {
        this.auth.redirectToDashboard(role);
      }
    }
  }

  loginWithGoogle(): void {
    // Google OAuth 2.0 — redirects to Google's sign-in page.
    // Requires GOOGLE_CLIENT_ID set in environment and a callback route configured
    // in Google Cloud Console (APIs & Services → Credentials).
    const clientId = (window as any).__GOOGLE_CLIENT_ID__ || '';
    if (!clientId) {
      this.error = 'Google login is not configured yet. Please use email and password to sign in.';
      return;
    }
    const redirectUri = encodeURIComponent(window.location.origin + '/auth/google/callback');
    const scope = encodeURIComponent('openid email profile');
    const url = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=${scope}&access_type=offline`;
    window.location.href = url;
  }

  onSubmit(): void {
    this.error = '';
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { email, password } = this.form.value;
    this.loading = true;

    this.auth.login(email, password).subscribe({
      next: (res) => {
        this.auth.redirectToDashboard(res.user.role);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || err.error?.message || 'Invalid email or password';
      }
    });
  }
}

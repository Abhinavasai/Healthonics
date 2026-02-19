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

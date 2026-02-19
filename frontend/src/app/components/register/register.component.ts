import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators, FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, FormsModule],
  templateUrl: './register.component.html',
  styleUrl: './register.component.scss'
})
export class RegisterComponent {
  form: FormGroup;
  error = '';
  loading = false;
  showPassword = false;
  emailFocused = false;
  passwordFocused = false;
  acceptTerms = false;

  roleOptions = [
    { value: 'patient', label: 'Patient', icon: '🧑‍⚕️' },
    { value: 'doctor', label: 'Doctor', icon: '👨‍⚕️' },
    { value: 'admin', label: 'Admin', icon: '🛡️' }
  ];

  constructor(
    private fb: FormBuilder,
    private auth: AuthService
  ) {
    this.form = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(6)]],
      role: ['patient', [Validators.required]]
    });
  }

  get passwordStrength(): number {
    const password = this.form.get('password')?.value || '';
    let strength = 0;
    if (password.length >= 6) strength++;
    if (password.length >= 8) strength++;
    if (/[A-Z]/.test(password) && /[a-z]/.test(password)) strength++;
    if (/[0-9]/.test(password) || /[^A-Za-z0-9]/.test(password)) strength++;
    return strength;
  }

  get passwordStrengthText(): string {
    const strength = this.passwordStrength;
    if (strength === 0) return '';
    if (strength === 1) return 'Weak';
    if (strength === 2) return 'Fair';
    if (strength === 3) return 'Good';
    return 'Strong';
  }

  onSubmit(): void {
    this.error = '';
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { email, password, role } = this.form.value;
    this.loading = true;

    this.auth.register(email, password, role).subscribe({
      next: (res) => {
        this.auth.redirectToDashboard(res.user.role);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || err.error?.message || 'Registration failed';
      }
    });
  }
}

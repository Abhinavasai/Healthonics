import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [RouterLink],
  template: `
    <div class="login-placeholder">
      <h1>Login</h1>
      <p>Login page coming in US-2.</p>
      <a routerLink="/register">Go to Register</a>
    </div>
  `,
  styles: [`
    .login-placeholder {
      max-width: 360px;
      margin: 2rem auto;
      padding: 2rem;
      text-align: center;
    }
    a { color: #1976d2; }
  `]
})
export class LoginComponent {}

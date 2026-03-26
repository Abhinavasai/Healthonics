import { Component } from '@angular/core';
import { CommonModule, TitleCasePipe } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss'
})
export class AppShellComponent {
  constructor(
    public auth: AuthService,
    private router: Router
  ) {}

  get user() {
    return this.auth.getUser();
  }

  get role() {
    return this.user?.role ?? 'patient';
  }

  logout(): void {
    this.auth.logout();
  }

  get navLinks(): { path: string; label: string; roles: string[] }[] {
    const all = [
      { path: '/patient/find-care', label: 'Find care', roles: ['patient'] },
      { path: '/patient/appointments', label: 'My Appointments', roles: ['patient'] },
      { path: '/doctor/availability', label: 'Availability', roles: ['doctor'] },
      { path: '/doctor/appointments', label: 'Appointment Queue', roles: ['doctor'] },
      { path: '/admin', label: 'Admin Dashboard', roles: ['admin'] }
    ];
    return all.filter((l) => l.roles.includes(this.role));
  }

}

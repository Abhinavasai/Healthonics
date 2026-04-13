import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, TitleCasePipe, AsyncPipe } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { Subscription, timer, switchMap } from 'rxjs';
import { AuthService } from '../../services/auth.service';
import { MessagingService } from '../../services/messaging.service';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe, AsyncPipe],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss'
})
export class AppShellComponent implements OnInit, OnDestroy {
  private unreadPoll?: Subscription;

  constructor(
    public auth: AuthService,
    public messaging: MessagingService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.unreadPoll = timer(0, 30_000)
      .pipe(switchMap(() => this.messaging.refreshUnread()))
      .subscribe();
  }

  ngOnDestroy(): void {
    this.unreadPoll?.unsubscribe();
  }

  get user() {
    return this.auth.getUser();
  }

  get role() {
    return this.user?.role ?? 'patient';
  }

  logout(): void {
    this.auth.logout();
  }

  /** False for routes with detail children (appointments/:id, messages/:threadId). */
  isExactNavPath(path: string): boolean {
    if (path.endsWith('/appointments') || path.endsWith('/messages') || path.endsWith('/my-files')) {
      return false;
    }
    return true;
  }

  get navLinks(): { path: string; label: string; roles: string[] }[] {
    const all = [
      { path: '/patient/find-care', label: 'Find care', roles: ['patient'] },
      { path: '/patient/my-files', label: 'My files', roles: ['patient'] },
      { path: '/patient/appointments', label: 'My Appointments', roles: ['patient'] },
      { path: '/patient/messages', label: 'Messages', roles: ['patient'] },
      { path: '/doctor/availability', label: 'Availability', roles: ['doctor'] },
      { path: '/doctor/appointments', label: 'Appointment Queue', roles: ['doctor'] },
      { path: '/doctor/messages', label: 'Messages', roles: ['doctor'] },
      { path: '/admin', label: 'Admin Dashboard', roles: ['admin'] }
    ];
    return all.filter((l) => l.roles.includes(this.role));
  }

}

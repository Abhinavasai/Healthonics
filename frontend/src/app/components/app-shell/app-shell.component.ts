import { CommonModule, DOCUMENT, TitleCasePipe, AsyncPipe } from '@angular/common';
import { Component, HostListener, Inject, OnDestroy, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { filter, Subscription, switchMap, timer } from 'rxjs';
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
  private navEndSub?: Subscription;

  /** Mobile drawer (Sprint 3 F02). */
  mobileNavOpen = false;

  constructor(
    public auth: AuthService,
    public messaging: MessagingService,
    private router: Router,
    @Inject(DOCUMENT) private doc: Document
  ) {}

  ngOnInit(): void {
    this.unreadPoll = timer(0, 30_000)
      .pipe(switchMap(() => this.messaging.refreshUnread()))
      .subscribe();

    this.navEndSub = this.router.events
      .pipe(filter((e): e is NavigationEnd => e instanceof NavigationEnd))
      .subscribe(() => this.closeMobileNav());
  }

  ngOnDestroy(): void {
    this.unreadPoll?.unsubscribe();
    this.navEndSub?.unsubscribe();
    this.doc.body.style.overflow = '';
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
    if (path.endsWith('/appointments') || path.endsWith('/messages')) {
      return false;
    }
    return true;
  }

  @HostListener('document:keydown.escape')
  onEscapeCloseNav(): void {
    if (this.mobileNavOpen) {
      this.closeMobileNav();
    }
  }

  toggleMobileNav(): void {
    this.mobileNavOpen = !this.mobileNavOpen;
    this.syncBodyScrollLock();
  }

  closeMobileNav(): void {
    this.mobileNavOpen = false;
    this.syncBodyScrollLock();
  }

  private syncBodyScrollLock(): void {
    this.doc.body.style.overflow = this.mobileNavOpen ? 'hidden' : '';
  }

  get navLinks(): { path: string; label: string; roles: string[] }[] {
    const all = [
      { path: '/patient/dashboard', label: 'Dashboard', roles: ['patient'] },
      { path: '/patient/find-care', label: 'Find care', roles: ['patient'] },
      { path: '/patient/appointments', label: 'My Appointments', roles: ['patient'] },
      { path: '/patient/messages', label: 'Messages', roles: ['patient'] },
      { path: '/patient/settings', label: 'Account', roles: ['patient'] },
      { path: '/doctor/dashboard', label: 'Dashboard', roles: ['doctor'] },
      { path: '/doctor/availability', label: 'Availability', roles: ['doctor'] },
      { path: '/doctor/appointments', label: 'Appointment Queue', roles: ['doctor'] },
      { path: '/doctor/messages', label: 'Messages', roles: ['doctor'] },
      { path: '/doctor/settings', label: 'Account', roles: ['doctor'] },
      { path: '/admin', label: 'Admin Dashboard', roles: ['admin'] }
    ];
    return all.filter((l) => l.roles.includes(this.role));
  }

}

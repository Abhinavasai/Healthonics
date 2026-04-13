import { CommonModule, DOCUMENT, TitleCasePipe } from '@angular/common';
import { Component, HostListener, Inject, OnDestroy, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { filter, Subscription } from 'rxjs';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, RouterModule, TitleCasePipe],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss'
})
export class AppShellComponent implements OnInit, OnDestroy {
  /** Mobile drawer (Sprint 3 F02 — Karthik): pairs with existing `.sidebar.open` styles. */
  mobileNavOpen = false;

  private navEndSub?: Subscription;

  constructor(
    public auth: AuthService,
    private router: Router,
    @Inject(DOCUMENT) private doc: Document
  ) {}

  ngOnInit(): void {
    this.navEndSub = this.router.events
      .pipe(filter((e): e is NavigationEnd => e instanceof NavigationEnd))
      .subscribe(() => this.closeMobileNav());
  }

  ngOnDestroy(): void {
    this.navEndSub?.unsubscribe();
    this.doc.body.style.overflow = '';
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
      { path: '/patient/settings', label: 'Account', roles: ['patient'] },
      { path: '/doctor/availability', label: 'Availability', roles: ['doctor'] },
      { path: '/doctor/appointments', label: 'Appointment Queue', roles: ['doctor'] },
      { path: '/doctor/settings', label: 'Account', roles: ['doctor'] },
      { path: '/admin', label: 'Admin Dashboard', roles: ['admin'] }
    ];
    return all.filter((l) => l.roles.includes(this.role));
  }

}

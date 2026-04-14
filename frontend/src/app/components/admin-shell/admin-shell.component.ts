import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-admin-shell',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  template: `
    <nav class="subnav">
      <a routerLink="/admin" routerLinkActive="active" [routerLinkActiveOptions]="{ exact: true }">Overview</a>
      <a routerLink="/admin/audit" routerLinkActive="active">Audit log</a>
      <a routerLink="/admin/knowledge" routerLinkActive="active">Knowledge</a>
    </nav>
    <router-outlet />
  `,
  styles: [
    `
      .subnav {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;
      }
      .subnav a {
        color: #94a3b8;
      }
      .subnav a.active {
        color: #22d3ee;
      }
    `
  ]
})
export class AdminShellComponent {}

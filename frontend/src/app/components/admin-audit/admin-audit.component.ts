import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AdminAuditService, AuditLogEntry } from '../../services/admin-audit.service';

@Component({
  selector: 'app-admin-audit',
  standalone: true,
  imports: [CommonModule],
  template: `
    <section data-cy="admin-audit-page">
      <h1>Audit log</h1>
      <p *ngIf="error" class="err">{{ error }}</p>
      <table *ngIf="entries.length" data-cy="admin-audit-table">
        <thead>
          <tr>
            <th>When</th>
            <th>Action</th>
            <th>Entity</th>
            <th>Detail</th>
          </tr>
        </thead>
        <tbody>
          <tr *ngFor="let e of entries" data-cy="admin-audit-row">
            <td>{{ e.created_at }}</td>
            <td>{{ e.action }}</td>
            <td>{{ e.entity_type }} {{ e.entity_id }}</td>
            <td>{{ e.detail }}</td>
          </tr>
        </tbody>
      </table>
      <p *ngIf="!error && !entries.length" data-cy="admin-audit-empty">No audit entries yet.</p>
    </section>
  `,
  styles: [
    `
      table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.85rem;
      }
      th,
      td {
        border-bottom: 1px solid rgba(148, 163, 184, 0.25);
        padding: 0.4rem;
        text-align: left;
      }
      .err {
        color: #f87171;
      }
    `
  ]
})
export class AdminAuditComponent implements OnInit {
  entries: AuditLogEntry[] = [];
  error: string | null = null;

  constructor(private api: AdminAuditService) {}

  ngOnInit(): void {
    this.api.list().subscribe({
      next: (r) => (this.entries = r.entries ?? []),
      error: () => (this.error = 'Could not load audit log')
    });
  }
}

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  AdminLifecycleKpis,
  AdminLifecycleSettings,
  AdminUserLifecycleService
} from '../../services/admin-user-lifecycle.service';

@Component({
  selector: 'app-admin-management',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './admin-management.component.html',
  styleUrl: './admin-management.component.scss'
})
export class AdminManagementComponent implements OnInit {
  loading = true;
  saving = false;
  error: string | null = null;
  success: string | null = null;
  settings: AdminLifecycleSettings | null = null;
  kpis: AdminLifecycleKpis | null = null;

  form = {
    new_user_window_days: 14,
    inactive_window_days: 30
  };

  constructor(private readonly api: AdminUserLifecycleService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getSettingsKpis().subscribe({
      next: (res) => {
        this.settings = res.settings;
        this.kpis = res.kpis;
        this.form.new_user_window_days = res.settings.new_user_window_days;
        this.form.inactive_window_days = res.settings.inactive_window_days;
        this.loading = false;
      },
      error: () => {
        this.error = 'Could not load lifecycle settings.';
        this.loading = false;
      }
    });
  }

  save(): void {
    this.saving = true;
    this.error = null;
    this.success = null;
    this.api
      .updateSettings({
        new_user_window_days: Number(this.form.new_user_window_days),
        inactive_window_days: Number(this.form.inactive_window_days)
      })
      .subscribe({
        next: (res) => {
          this.settings = res.settings;
          this.kpis = res.kpis;
          this.success = 'Settings updated.';
          this.saving = false;
        },
        error: () => {
          this.error = 'Could not update settings.';
          this.saving = false;
        }
      });
  }
}

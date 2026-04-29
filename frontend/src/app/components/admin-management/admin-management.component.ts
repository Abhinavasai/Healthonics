import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  AdminLifecycleKpis,
  AdminLifecycleSettings,
  AdminLifecycleUser,
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
  userBusy = false;
  error: string | null = null;
  success: string | null = null;
  settings: AdminLifecycleSettings | null = null;
  kpis: AdminLifecycleKpis | null = null;
  users: AdminLifecycleUser[] = [];
  selectedUserId = '';

  form = {
    new_user_window_days: 14,
    inactive_window_days: 30
  };
  createForm = {
    email: '',
    password: '',
    role: 'patient' as 'patient' | 'doctor' | 'admin'
  };
  editForm = {
    email: '',
    role: 'patient' as 'patient' | 'doctor' | 'admin'
  };
  resetForm = {
    newPassword: ''
  };
  settingsGuard = {
    reason: '',
    confirm: ''
  };
  deactivateGuard = {
    reason: '',
    confirm: ''
  };
  resetGuard = {
    reason: '',
    confirm: ''
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
        this.api.listUsers().subscribe({
          next: (usersRes) => {
            this.users = usersRes.users ?? [];
            if (this.users.length > 0 && !this.selectedUserId) {
              this.selectUser(this.users[0].id);
            }
            this.loading = false;
          },
          error: () => {
            this.error = 'Could not load users.';
            this.loading = false;
          }
        });
      },
      error: () => {
        this.error = 'Could not load lifecycle settings.';
        this.loading = false;
      }
    });
  }

  save(): void {
    if (!this.validGuardrail(this.settingsGuard.reason, this.settingsGuard.confirm)) {
      this.error = 'Settings update requires a reason and CONFIRM.';
      return;
    }
    this.saving = true;
    this.error = null;
    this.success = null;
    this.api
      .updateSettings({
        new_user_window_days: Number(this.form.new_user_window_days),
        inactive_window_days: Number(this.form.inactive_window_days),
        reason: this.settingsGuard.reason.trim(),
        confirm: this.settingsGuard.confirm.trim().toUpperCase()
      })
      .subscribe({
        next: (res) => {
          this.settings = res.settings;
          this.kpis = res.kpis;
          this.success = 'Settings updated.';
          this.saving = false;
        },
        error: () => {
          this.error = 'Could not update settings. Reason + CONFIRM are required.';
          this.saving = false;
        }
      });
  }

  selectUser(id: string): void {
    this.selectedUserId = id;
    const u = this.users.find((x) => x.id === id);
    if (!u) {
      return;
    }
    this.editForm.email = u.email;
    this.editForm.role = u.role;
    this.resetForm.newPassword = '';
  }

  createUser(): void {
    this.userBusy = true;
    this.error = null;
    this.success = null;
    this.api
      .createUser({
        email: this.createForm.email.trim().toLowerCase(),
        password: this.createForm.password,
        role: this.createForm.role
      })
      .subscribe({
        next: () => {
          this.success = 'User created.';
          this.createForm.email = '';
          this.createForm.password = '';
          this.refreshUsers();
        },
        error: () => {
          this.userBusy = false;
          this.error = 'Could not create user.';
        }
      });
  }

  updateSelectedUser(): void {
    if (!this.selectedUserId) {
      return;
    }
    this.userBusy = true;
    this.error = null;
    this.success = null;
    this.api
      .updateUser(this.selectedUserId, {
        email: this.editForm.email.trim().toLowerCase(),
        role: this.editForm.role
      })
      .subscribe({
        next: () => {
          this.success = 'User updated.';
          this.refreshUsers();
        },
        error: () => {
          this.userBusy = false;
          this.error = 'Could not update user.';
        }
      });
  }

  deactivateSelectedUser(): void {
    if (!this.selectedUserId) {
      return;
    }
    if (!this.validGuardrail(this.deactivateGuard.reason, this.deactivateGuard.confirm)) {
      this.error = 'Deactivate requires a reason and CONFIRM.';
      return;
    }
    this.userBusy = true;
    this.error = null;
    this.success = null;
    this.api
      .deactivateUser(this.selectedUserId, {
        reason: this.deactivateGuard.reason.trim(),
        confirm: this.deactivateGuard.confirm.trim().toUpperCase()
      })
      .subscribe({
      next: () => {
        this.success = 'User deactivated.';
        this.deactivateGuard = { reason: '', confirm: '' };
        this.refreshUsers();
      },
      error: () => {
        this.userBusy = false;
        this.error = 'Could not deactivate user.';
      }
      });
  }

  resetSelectedPassword(): void {
    if (!this.selectedUserId || !this.resetForm.newPassword.trim()) {
      return;
    }
    if (!this.validGuardrail(this.resetGuard.reason, this.resetGuard.confirm)) {
      this.error = 'Reset password requires a reason and CONFIRM.';
      return;
    }
    this.userBusy = true;
    this.error = null;
    this.success = null;
    this.api
      .resetPassword(this.selectedUserId, this.resetForm.newPassword, {
        reason: this.resetGuard.reason.trim(),
        confirm: this.resetGuard.confirm.trim().toUpperCase()
      })
      .subscribe({
      next: () => {
        this.success = 'Password reset complete.';
        this.resetForm.newPassword = '';
        this.resetGuard = { reason: '', confirm: '' };
        this.userBusy = false;
      },
      error: () => {
        this.userBusy = false;
        this.error = 'Could not reset password.';
      }
      });
  }

  private refreshUsers(): void {
    this.api.listUsers().subscribe({
      next: (res) => {
        this.users = res.users ?? [];
        const stillExists = this.users.some((u) => u.id === this.selectedUserId);
        if (!stillExists && this.users.length > 0) {
          this.selectUser(this.users[0].id);
        } else if (stillExists) {
          this.selectUser(this.selectedUserId);
        }
        this.userBusy = false;
      },
      error: () => {
        this.userBusy = false;
        this.error = 'Could not refresh users.';
      }
    });
  }

  private validGuardrail(reason: string, confirm: string): boolean {
    return reason.trim().length >= 1 && confirm.trim().toUpperCase() === 'CONFIRM';
  }
}

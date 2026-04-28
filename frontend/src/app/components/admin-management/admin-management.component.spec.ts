import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { AdminManagementComponent } from './admin-management.component';

describe('AdminManagementComponent', () => {
  let fixture: ComponentFixture<AdminManagementComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminManagementComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();

    fixture = TestBed.createComponent(AdminManagementComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();

    httpMock.expectOne('/api/admin/user-lifecycle/settings-kpis').flush({
      settings: {
        new_user_window_days: 14,
        inactive_window_days: 30,
        updated_at: 'now',
        updated_by: ''
      },
      kpis: {
        total_users: 2,
        total_patients: 1,
        total_doctors: 1,
        total_admins: 0,
        new_users_in_window: 1,
        inactive_users_count: 0
      }
    });
    httpMock.expectOne('/api/admin/user-lifecycle/users').flush({
      users: [
        {
          id: 'u1',
          email: 'patient@demo.com',
          role: 'patient',
          is_active: true,
          created_at: 'now',
          deactivated_at: '',
          deactivated_by: ''
        }
      ]
    });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('renders users table', () => {
    expect((fixture.nativeElement as HTMLElement).querySelector('[data-cy="admin-users-table"]')).toBeTruthy();
  });

  it('creates a user', () => {
    fixture.componentInstance.createForm.email = 'new@demo.com';
    fixture.componentInstance.createForm.password = 'strongpass';
    fixture.componentInstance.createForm.role = 'doctor';
    fixture.componentInstance.createUser();

    const createReq = httpMock.expectOne('/api/admin/user-lifecycle/users');
    expect(createReq.request.method).toBe('POST');
    createReq.flush({
      id: 'u2',
      email: 'new@demo.com',
      role: 'doctor',
      is_active: true
    });
    httpMock.expectOne('/api/admin/user-lifecycle/users').flush({
      users: [
        {
          id: 'u1',
          email: 'patient@demo.com',
          role: 'patient',
          is_active: true,
          created_at: 'now',
          deactivated_at: '',
          deactivated_by: ''
        },
        {
          id: 'u2',
          email: 'new@demo.com',
          role: 'doctor',
          is_active: true,
          created_at: 'now',
          deactivated_at: '',
          deactivated_by: ''
        }
      ]
    });
    expect(fixture.componentInstance.success).toBe('User created.');
  });

  it('updates selected user', () => {
    fixture.componentInstance.selectUser('u1');
    fixture.componentInstance.editForm.email = 'updated@demo.com';
    fixture.componentInstance.editForm.role = 'doctor';
    fixture.componentInstance.updateSelectedUser();

    const updateReq = httpMock.expectOne('/api/admin/user-lifecycle/users/u1');
    expect(updateReq.request.method).toBe('PUT');
    updateReq.flush({
      id: 'u1',
      email: 'updated@demo.com',
      role: 'doctor',
      is_active: true
    });
    httpMock.expectOne('/api/admin/user-lifecycle/users').flush({
      users: [
        {
          id: 'u1',
          email: 'updated@demo.com',
          role: 'doctor',
          is_active: true,
          created_at: 'now',
          deactivated_at: '',
          deactivated_by: ''
        }
      ]
    });
    expect(fixture.componentInstance.success).toBe('User updated.');
  });

  it('deactivates selected user', () => {
    fixture.componentInstance.selectUser('u1');
    fixture.componentInstance.deactivateSelectedUser();
    const deactivateReq = httpMock.expectOne('/api/admin/user-lifecycle/users/u1/deactivate');
    expect(deactivateReq.request.method).toBe('PATCH');
    deactivateReq.flush({
      id: 'u1',
      status: 'deactivated'
    });
    httpMock.expectOne('/api/admin/user-lifecycle/users').flush({
      users: [
        {
          id: 'u1',
          email: 'patient@demo.com',
          role: 'patient',
          is_active: false,
          created_at: 'now',
          deactivated_at: 'now',
          deactivated_by: 'admin'
        }
      ]
    });
    expect(fixture.componentInstance.success).toBe('User deactivated.');
  });

  it('resets selected user password', () => {
    fixture.componentInstance.selectUser('u1');
    fixture.componentInstance.resetForm.newPassword = 'newpass123';
    fixture.componentInstance.resetSelectedPassword();
    const resetReq = httpMock.expectOne('/api/admin/user-lifecycle/users/u1/reset-password');
    expect(resetReq.request.method).toBe('POST');
    resetReq.flush({
      id: 'u1',
      password_reset: true
    });
    expect(fixture.componentInstance.resetForm.newPassword).toBe('');
    expect(fixture.componentInstance.success).toBe('Password reset complete.');
  });
});


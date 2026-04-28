import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { AdminNotificationsLogComponent } from './admin-notifications-log.component';

describe('AdminNotificationsLogComponent', () => {
  let fixture: ComponentFixture<AdminNotificationsLogComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AdminNotificationsLogComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();

    fixture = TestBed.createComponent(AdminNotificationsLogComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty state', () => {
    httpMock.expectOne('/api/admin/notifications/summary').flush({
      pending_count: 0,
      sent_count: 0,
      failed_count: 0
    });
    httpMock.expectOne('/api/admin/notifications').flush({ notifications: [] });
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="admin-notifications-empty"]')
    ).toBeTruthy();
  });

  it('renders table rows', () => {
    httpMock.expectOne('/api/admin/notifications/summary').flush({
      pending_count: 1,
      sent_count: 0,
      failed_count: 1
    });
    httpMock.expectOne('/api/admin/notifications').flush({
      notifications: [
        {
          id: '1',
          user_id: 'u-1',
          title: 'Test',
          body: 'Hello',
          channel: 'in_app',
          provider: 'in_app',
          attempts: 1,
          last_error: '',
          status: 'sent',
          created_at: 'now'
        }
      ]
    });
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelectorAll('[data-cy="admin-notifications-row"]')
        .length
    ).toBe(1);
  });

  it('retries failed notifications', () => {
    httpMock.expectOne('/api/admin/notifications/summary').flush({
      pending_count: 1,
      sent_count: 0,
      failed_count: 1
    });
    httpMock.expectOne('/api/admin/notifications').flush({
      notifications: [
        {
          id: '1',
          user_id: 'u-1',
          title: 'Test',
          body: 'Hello',
          channel: 'sms',
          provider: 'twilio',
          attempts: 2,
          last_error: 'twilio timeout',
          next_retry_at: 'later',
          status: 'failed',
          created_at: 'now'
        }
      ]
    });
    fixture.detectChanges();
    (fixture.nativeElement as HTMLElement).querySelector('[data-cy="retry-failed-notification"]')?.dispatchEvent(
      new Event('click')
    );
    fixture.detectChanges();
    httpMock.expectOne('/api/admin/notifications/1/retry').flush({ id: '1', status: 'pending' });
    httpMock.expectOne('/api/admin/notifications/summary').flush({
      pending_count: 1,
      sent_count: 0,
      failed_count: 0
    });
    httpMock.expectOne('/api/admin/notifications').flush({ notifications: [] });
    expect(fixture.componentInstance.retryingId).toBeNull();
  });

  it('renders provider health cards', () => {
    httpMock.expectOne('/api/admin/notifications/summary').flush({
      pending_count: 1,
      sent_count: 1,
      failed_count: 1
    });
    httpMock.expectOne('/api/admin/notifications').flush({
      notifications: [
        {
          id: '1',
          user_id: 'u-1',
          title: 'Email delivered',
          body: 'ok',
          channel: 'email',
          provider: 'sendgrid',
          attempts: 1,
          last_error: '',
          status: 'sent',
          created_at: 'now'
        },
        {
          id: '2',
          user_id: 'u-2',
          title: 'SMS failed',
          body: 'retrying',
          channel: 'sms',
          provider: 'twilio',
          attempts: 3,
          last_error: 'invalid number',
          status: 'failed',
          created_at: 'now'
        }
      ]
    });
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="provider-health-cards"]')
    ).toBeTruthy();
  });
});


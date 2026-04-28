import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { NotificationsInboxComponent } from './notifications-inbox.component';

describe('NotificationsInboxComponent', () => {
  let fixture: ComponentFixture<NotificationsInboxComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotificationsInboxComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();

    fixture = TestBed.createComponent(NotificationsInboxComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/notifications/preferences').flush({
      preferences: [
        {
          category: 'appointment_reminders',
          enabled: true,
          email_enabled: true,
          sms_enabled: false,
          in_app_enabled: true
        },
        {
          category: 'announcements',
          enabled: true,
          email_enabled: true,
          sms_enabled: false,
          in_app_enabled: true
        }
      ]
    });
    httpMock.expectOne('/api/notifications').flush({ notifications: [] });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty state', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="notifications-empty"]')
    ).toBeTruthy();
  });

  it('shows notification preferences section', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="notification-preferences"]')
    ).toBeTruthy();
    expect(
      (fixture.nativeElement as HTMLElement).querySelectorAll('[data-cy="notification-pref-row"]')
        .length
    ).toBe(2);
  });

  it('supports bulk disable all controls', () => {
    (
      (fixture.nativeElement as HTMLElement).querySelector(
        '[data-cy="pref-bulk-disable-all"]'
      ) as HTMLButtonElement
    ).click();
    fixture.detectChanges();
    expect(fixture.componentInstance.preferences.every((p) => !p.enabled)).toBeTrue();
    expect(fixture.componentInstance.preferences.every((p) => !p.email_enabled)).toBeTrue();
    expect(fixture.componentInstance.preferences.every((p) => !p.sms_enabled)).toBeTrue();
    expect(fixture.componentInstance.preferences.every((p) => !p.in_app_enabled)).toBeTrue();
  });

  it('saves updated preferences payload', () => {
    (
      (fixture.nativeElement as HTMLElement).querySelector(
        '[data-cy="pref-bulk-sms-on"]'
      ) as HTMLButtonElement
    ).click();
    fixture.detectChanges();

    (
      (fixture.nativeElement as HTMLElement).querySelector(
        '[data-cy="notification-pref-save"]'
      ) as HTMLButtonElement
    ).click();

    const req = httpMock.expectOne('/api/notifications/preferences');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body.preferences.length).toBe(2);
    expect(req.request.body.preferences.every((p: { sms_enabled: boolean }) => p.sms_enabled)).toBeTrue();
    req.flush({ updated: 2 });
  });
});

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
    httpMock.expectOne('/api/notifications').flush({ notifications: [] });
    fixture.detectChanges();
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="admin-notifications-empty"]')
    ).toBeTruthy();
  });

  it('renders table rows', () => {
    httpMock.expectOne('/api/notifications').flush({
      notifications: [
        {
          id: '1',
          title: 'Test',
          body: 'Hello',
          channel: 'in_app',
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
});


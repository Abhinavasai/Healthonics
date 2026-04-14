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
    httpMock.expectOne('/api/notifications').flush({ notifications: [] });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty state', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="notifications-empty"]')
    ).toBeTruthy();
  });
});

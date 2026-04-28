import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { PatientDashboardComponent } from './patient-dashboard.component';

describe('PatientDashboardComponent', () => {
  let fixture: ComponentFixture<PatientDashboardComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PatientDashboardComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();
    fixture = TestBed.createComponent(PatientDashboardComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/patient/dashboard/summary').flush({
      role: 'patient',
      upcoming_or_active_count: 2,
      pending_requests: 1,
      aggregations: {
        unread_messages: 3,
        active_prescriptions: 1,
        documents_uploaded: 4
      },
      alerts: [{ severity: 'info', code: 'unread_messages', message: 'Unread messages in your inbox.', count: 3 }]
    });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('renders stats, chart, bars, and insights', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="patient-dashboard-stats"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="patient-dashboard-chart"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="patient-dashboard-bars"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="patient-dashboard-insights"]')).toBeTruthy();
    expect(el.textContent).toContain('2');
  });
});

import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { DoctorDashboardComponent } from './doctor-dashboard.component';

describe('DoctorDashboardComponent', () => {
  let fixture: ComponentFixture<DoctorDashboardComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DoctorDashboardComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    }).compileComponents();
    fixture = TestBed.createComponent(DoctorDashboardComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/doctor/dashboard/summary').flush({
      role: 'doctor',
      appointments_today: 2,
      pending_queue: 1,
      aggregations: {
        unread_messages: 4,
        appointments_this_week: 12
      },
      alerts: [
        { severity: 'warning', code: 'pending_queue', message: 'Appointment requests need your review.', count: 1 }
      ]
    });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('renders stats, workload ring, bars, and alerts', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-dashboard-stats"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="doctor-dashboard-workload-ring"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="doctor-dashboard-workload-bars"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="doctor-dashboard-alerts"]')).toBeTruthy();
  });
});

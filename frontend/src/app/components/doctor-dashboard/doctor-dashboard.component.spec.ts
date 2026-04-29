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
        appointments_this_week: 12,
        critical_open_count: 1
      },
      alerts: [
        { severity: 'warning', code: 'pending_queue', message: 'Appointment requests need your review.', count: 1 }
      ],
      critical_escalations: [
        {
          id: 'esc-1',
          document_id: 'doc-1',
          patient_id: 'pat-1',
          patient_email: 'p@h.com',
          filename: 'lab.pdf',
          tier: 2,
          status: 'open',
          rule_code: 'critical_summary_keyword',
          message: 'Tier 2 critical result: lab.pdf for patient p@h.com requires review.',
          first_detected_at: '2026-04-28T20:00:00Z',
          next_escalation_at: '2026-04-28T20:15:00Z'
        }
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
    expect(el.querySelector('[data-cy="doctor-critical-escalations"]')).toBeTruthy();
  });

  it('acknowledges an open critical escalation', () => {
    const el = fixture.nativeElement as HTMLElement;
    const btn = el.querySelector('[data-cy="ack-critical-esc-1"]') as HTMLButtonElement;
    expect(btn).toBeTruthy();
    btn.click();

    httpMock.expectOne('/api/doctor/critical-escalations/esc-1/ack').flush({ id: 'esc-1', status: 'acknowledged' });
    fixture.detectChanges();

    expect(fixture.componentInstance.data?.critical_escalations?.[0]?.status).toBe('acknowledged');
  });
});

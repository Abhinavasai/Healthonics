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
      pending_requests: 1
    });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('renders stats', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="patient-dashboard-stats"]')).toBeTruthy();
    expect(el.textContent).toContain('2');
  });
});

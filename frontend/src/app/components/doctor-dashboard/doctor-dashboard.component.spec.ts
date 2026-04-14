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
      appointments_today: 1,
      pending_queue: 0
    });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('renders stats', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="doctor-dashboard-stats"]')
    ).toBeTruthy();
  });
});

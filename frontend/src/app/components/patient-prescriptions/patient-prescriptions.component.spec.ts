import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { PatientPrescriptionsComponent } from './patient-prescriptions.component';
import { AuthService } from '../../services/auth.service';

describe('PatientPrescriptionsComponent', () => {
  let fixture: ComponentFixture<PatientPrescriptionsComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PatientPrescriptionsComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: AuthService,
          useValue: { getUser: () => ({ id: 'p1', email: 'a@b.com', role: 'patient' }) }
        }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(PatientPrescriptionsComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/patients/p1/prescriptions').flush({ prescriptions: [] });
    fixture.detectChanges();
  });

  afterEach(() => httpMock.verify());

  it('shows empty state', () => {
    expect(
      (fixture.nativeElement as HTMLElement).querySelector('[data-cy="patient-prescriptions-empty"]')
    ).toBeTruthy();
  });

  it('renders active and history sections', () => {
    fixture = TestBed.createComponent(PatientPrescriptionsComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/patients/p1/prescriptions').flush({
      prescriptions: [
        {
          id: 'rx1',
          patient_id: 'p1',
          doctor_id: 'd1',
          medication_name: 'Atenolol',
          dosage: '25mg',
          frequency: 'daily',
          duration_days: 30,
          instructions: 'after food',
          status: 'active',
          created_at: '2026-04-20T10:00:00Z'
        },
        {
          id: 'rx2',
          patient_id: 'p1',
          doctor_id: 'd1',
          medication_name: 'Statin',
          dosage: '10mg',
          frequency: 'nightly',
          duration_days: 14,
          instructions: '',
          status: 'revoked',
          created_at: '2026-03-20T10:00:00Z'
        }
      ]
    });
    fixture.detectChanges();

    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelectorAll('[data-cy="patient-prescriptions-row"]').length).toBe(1);
    expect(el.querySelectorAll('[data-cy="patient-prescriptions-history-row"]').length).toBe(1);
    expect(el.querySelector('[data-cy="patient-prescriptions-summary"]')).toBeTruthy();
  });

  it('toggles reminder preference for active prescription', () => {
    fixture = TestBed.createComponent(PatientPrescriptionsComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/patients/p1/prescriptions').flush({
      prescriptions: [
        {
          id: 'rx1',
          patient_id: 'p1',
          doctor_id: 'd1',
          medication_name: 'Atenolol',
          dosage: '25mg',
          frequency: 'daily',
          duration_days: 30,
          instructions: 'after food',
          status: 'active',
          created_at: '2026-04-20T10:00:00Z'
        }
      ]
    });
    fixture.detectChanges();

    const el = fixture.nativeElement as HTMLElement;
    const statusBefore = el.querySelector('.rem-status')?.textContent ?? '';
    expect(statusBefore).toContain('On');

    const btn = el.querySelector('[data-cy="patient-prescriptions-reminder-toggle"]') as HTMLButtonElement;
    btn.click();
    fixture.detectChanges();

    const statusAfter = (fixture.nativeElement as HTMLElement).querySelector('.rem-status')?.textContent ?? '';
    expect(statusAfter).toContain('Off');
  });
});

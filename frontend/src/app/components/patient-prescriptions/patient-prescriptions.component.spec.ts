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
});

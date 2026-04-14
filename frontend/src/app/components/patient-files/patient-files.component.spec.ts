import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { PatientFilesComponent } from './patient-files.component';
import { AuthService } from '../../services/auth.service';

describe('PatientFilesComponent', () => {
  let fixture: ComponentFixture<PatientFilesComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PatientFilesComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: AuthService,
          useValue: {
            getUser: () => ({ id: 'p1', email: 'a@b.com', role: 'patient' })
          }
        }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(PatientFilesComponent);
    httpMock = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
    httpMock.expectOne('/api/patients/p1/files').flush({ files: [] });
    fixture.detectChanges();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('shows empty state', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="patient-files-empty"]')).toBeTruthy();
  });
});

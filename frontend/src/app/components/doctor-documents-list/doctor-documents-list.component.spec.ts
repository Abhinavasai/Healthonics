import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { DoctorDocumentsListComponent } from './doctor-documents-list.component';
import { DoctorDocumentsService } from '../../services/doctor-documents.service';

describe('DoctorDocumentsListComponent', () => {
  let fixture: ComponentFixture<DoctorDocumentsListComponent>;
  const serviceMock = {
    list: jasmine.createSpy('list').and.returnValue(
      of([
        {
          id: 'd1',
          filename: 'x.pdf',
          patient_email: 'p@test.com',
          created_at: '2026-01-01T00:00:00Z',
          summary_status: 'ready'
        }
      ])
    )
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DoctorDocumentsListComponent],
      providers: [provideRouter([]), { provide: DoctorDocumentsService, useValue: serviceMock }]
    }).compileComponents();

    fixture = TestBed.createComponent(DoctorDocumentsListComponent);
    fixture.detectChanges();
  });

  it('renders summary status pill', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-documents-summary-status"]')?.textContent).toContain('Summary ready');
  });
});

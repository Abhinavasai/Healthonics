import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { DoctorDocumentDetailComponent } from './doctor-document-detail.component';
import { DoctorDocumentsService } from '../../services/doctor-documents.service';

describe('DoctorDocumentDetailComponent', () => {
  let fixture: ComponentFixture<DoctorDocumentDetailComponent>;
  let component: DoctorDocumentDetailComponent;
  const serviceMock = {
    getDetail: jasmine.createSpy('getDetail').and.returnValue(
      of({
        id: 'd1',
        filename: 'lab.pdf',
        patient_id: 'p1',
        patient_email: 'p@test.com',
        size_bytes: 1024,
        content_type: 'application/pdf',
        created_at: '2026-01-01T00:00:00Z',
        summary: null,
        summary_status: 'pending'
      })
    ),
    requestSummary: jasmine.createSpy('requestSummary').and.returnValue(of({ status: 'pending', job_id: 'j1' }))
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DoctorDocumentDetailComponent],
      providers: [
        provideRouter([]),
        {
          provide: ActivatedRoute,
          useValue: { paramMap: of(convertToParamMap({ documentId: 'd1' })) }
        },
        { provide: DoctorDocumentsService, useValue: serviceMock }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(DoctorDocumentDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('renders summary status pill', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-document-summary-status"]')?.textContent).toContain(
      'Summary pending'
    );
  });

  it('renders preview section', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-document-preview-section"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="doctor-document-preview-placeholder"]')).toBeTruthy();
  });
});

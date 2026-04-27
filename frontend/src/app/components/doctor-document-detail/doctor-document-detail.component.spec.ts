import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
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

  it('shows Generating on primary button when summary is pending', () => {
    const btn = (fixture.nativeElement as HTMLElement).querySelector('[data-cy="doctor-document-summarize"]');
    expect(btn?.textContent?.trim()).toBe('Generating…');
  });

  it('sets summaryRequestError and shows inline retry when requestSummary fails', () => {
    serviceMock.requestSummary.and.returnValue(throwError(() => ({ error: { error: 'Rate limited' } })));
    const comp = fixture.componentInstance;
    comp.requestSummary();
    expect(comp.summaryRequestError).toBe('Rate limited');
    fixture.detectChanges();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-document-summary-request-error"]')?.textContent).toContain('Rate limited');
    expect(el.querySelector('[data-cy="doctor-document-summary-retry-from-error"]')).toBeTruthy();
  });
});

describe('DoctorDocumentDetailComponent (failed summary)', () => {
  let fixture: ComponentFixture<DoctorDocumentDetailComponent>;
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
        summary_status: 'failed' as const,
        summary_error: 'Extraction failed'
      })
    ),
    requestSummary: jasmine.createSpy('requestSummary').and.returnValue(of({ status: 'pending' }))
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
    fixture.detectChanges();
  });

  it('renders failed state with backend error and retry control', () => {
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-document-summary-failed"]')).toBeTruthy();
    expect(el.querySelector('[data-cy="doctor-document-summary-error-text"]')?.textContent).toContain('Extraction failed');
    expect(el.querySelector('[data-cy="doctor-document-summary-retry"]')?.textContent?.trim()).toBe('Try again');
    const primary = el.querySelector('[data-cy="doctor-document-summarize"]');
    expect(primary?.textContent?.trim()).toBe('Try again');
  });
});

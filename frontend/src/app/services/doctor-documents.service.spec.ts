import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DoctorDocumentsService } from './doctor-documents.service';

describe('DoctorDocumentsService', () => {
  let service: DoctorDocumentsService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [DoctorDocumentsService]
    });
    service = TestBed.inject(DoctorDocumentsService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('lists documents', (done) => {
    service.list().subscribe((rows) => {
      expect(rows.length).toBe(1);
      expect(rows[0].patient_email).toBe('p@test.com');
      done();
    });
    const req = httpMock.expectOne('/api/doctor/documents');
    req.flush({
      documents: [
        {
          id: 'a',
          filename: 'x.pdf',
          patient_email: 'p@test.com',
          created_at: '2026-01-01T00:00:00Z',
          summary_status: 'ready'
        }
      ]
    });
  });

  it('getDetail returns document', (done) => {
    service.getDetail('doc-1').subscribe((d) => {
      expect(d?.filename).toBe('lab.pdf');
      expect(d?.summary_status).toBe('ready');
      done();
    });
    const req = httpMock.expectOne('/api/doctor/documents/doc-1');
    req.flush({
      id: 'doc-1',
      filename: 'lab.pdf',
      patient_id: 'p1',
      patient_email: 'p@test.com',
      size_bytes: 100,
      content_type: 'application/pdf',
      created_at: '2026-01-01T00:00:00Z',
      summary: 'Stub summary',
      summary_status: 'ready'
    });
  });

  it('requestSummary posts summarize', () => {
    service.requestSummary('doc-1').subscribe();
    const req = httpMock.expectOne('/api/doctor/documents/doc-1/summarize');
    expect(req.request.method).toBe('POST');
    req.flush({ status: 'pending' });
  });
});

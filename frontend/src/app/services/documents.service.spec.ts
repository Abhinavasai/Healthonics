import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DocumentsService } from './documents.service';

describe('DocumentsService', () => {
  let service: DocumentsService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [DocumentsService]
    });
    service = TestBed.inject(DocumentsService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('lists documents', (done) => {
    service.list().subscribe((docs) => {
      expect(docs.length).toBe(1);
      expect(docs[0].filename).toBe('lab.pdf');
      done();
    });
    const req = httpMock.expectOne('/api/documents');
    expect(req.request.method).toBe('GET');
    req.flush({ documents: [{ id: 'a', filename: 'lab.pdf', created_at: '2026-01-01T00:00:00Z' }] });
  });

  it('upload posts multipart and completes with document', (done) => {
    const file = new File(['x'], 'test.pdf', { type: 'application/pdf' });

    service.upload(file).subscribe({
      next: (ev) => {
        if (ev.type === 'done') {
          expect(ev.doc.id).toBe('new-id');
        }
      },
      complete: () => done()
    });

    const req = httpMock.expectOne('/api/documents');
    expect(req.request.method).toBe('POST');
    req.flush(
      { id: 'new-id', filename: 'test.pdf', created_at: '2026-01-01T00:00:00Z' },
      { status: 201, statusText: 'Created' }
    );
  });

  it('downloadBlob requests file stream', () => {
    service.downloadBlob('doc-1').subscribe((b) => expect(b.size).toBe(2));
    const req = httpMock.expectOne('/api/documents/doc-1/download');
    expect(req.request.method).toBe('GET');
    expect(req.request.responseType).toBe('blob');
    req.flush(new Blob(['ok']));
  });
});

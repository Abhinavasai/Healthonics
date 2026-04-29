import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { FhirBoundaryService } from './fhir-boundary.service';
import { ApiContract } from './api-contract';
import type { FhirBundle } from './api-types';

describe('FhirBoundaryService', () => {
  let service: FhirBoundaryService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [FhirBoundaryService, provideHttpClient(), provideHttpClientTesting()]
    });
    service = TestBed.inject(FhirBoundaryService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('exports bundle from admin FHIR endpoint', () => {
    service.exportBundle().subscribe((res) => {
      expect(res.bundle.resourceType).toBe('Bundle');
      expect(res.bundle.entry.length).toBe(1);
    });
    const req = httpMock.expectOne(ApiContract.admin.fhirExport);
    expect(req.request.method).toBe('GET');
    req.flush({ bundle: { resourceType: 'Bundle', type: 'collection', entry: [{ resource: { resourceType: 'Patient' } }] } });
  });

  it('imports sample bundle with apply=false by default', () => {
    const sample: FhirBundle = {
      resourceType: 'Bundle',
      type: 'collection',
      entry: [
        { resource: { resourceType: 'Patient', id: 'p1', identifier: [{ value: 'patient@healthonyx.demo' }] } },
        { resource: { resourceType: 'Observation', id: 'obs-1', subject: { reference: 'Patient/p1' }, valueString: 'Normal' } }
      ]
    };
    service.importBundle(sample).subscribe((res) => {
      expect(res.validated).toBeTrue();
      expect(res.applied).toBeFalse();
      expect(res.patients_count).toBe(1);
    });
    const req = httpMock.expectOne(ApiContract.admin.fhirImport);
    expect(req.request.method).toBe('POST');
    expect(req.request.body.apply).toBeFalse();
    req.flush({
      validated: true,
      applied: false,
      patients_count: 1,
      encounters_count: 0,
      observations_count: 1,
      medications_count: 0
    });
  });
});


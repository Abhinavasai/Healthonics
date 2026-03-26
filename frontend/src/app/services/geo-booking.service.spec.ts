import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { GeoBookingService } from './geo-booking.service';

describe('GeoBookingService', () => {
  let service: GeoBookingService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });

    service = TestBed.inject(GeoBookingService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('geocodeSearch should call /api/geocode with q and limit', () => {
    service.geocodeSearch('Gainesville, FL', 3).subscribe((res) => {
      expect(res.results.length).toBe(1);
      expect(res.results[0].display_name).toBe('Gainesville, FL');
    });

    const req = httpMock.expectOne((r) => r.url === '/api/geocode' && r.method === 'GET');
    expect(req.request.params.get('q')).toBe('Gainesville, FL');
    expect(req.request.params.get('limit')).toBe('3');

    req.flush({
      results: [{ lat: 29.6516, lng: -82.3248, display_name: 'Gainesville, FL' }],
    });
  });

  it('hospitalsNear should call /api/hospitals/near with lat/lng/radius_km', () => {
    service.hospitalsNear(29.6516, -82.3248, 25).subscribe((res) => {
      expect(res.hospitals.length).toBe(1);
      expect(res.hospitals[0].name).toBe('UF Health Shands Hospital');
    });

    const req = httpMock.expectOne((r) => r.url === '/api/hospitals/near' && r.method === 'GET');
    expect(req.request.params.get('lat')).toBe('29.6516');
    expect(req.request.params.get('lng')).toBe('-82.3248');
    expect(req.request.params.get('radius_km')).toBe('25');

    req.flush({
      hospitals: [
        {
          id: '11111111-1111-1111-1111-111111111111',
          name: 'UF Health Shands Hospital',
          city: 'Gainesville',
          region: 'Florida',
          latitude: 29.6516,
          longitude: -82.3248,
          distance_km: 1.23,
        },
      ],
    });
  });
});


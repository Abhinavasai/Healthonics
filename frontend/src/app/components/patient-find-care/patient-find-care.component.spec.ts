import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { PatientFindCareComponent } from './patient-find-care.component';
import { GeoBookingService } from '../../services/geo-booking.service';

describe('PatientFindCareComponent', () => {
  let fixture: ComponentFixture<PatientFindCareComponent>;
  let component: PatientFindCareComponent;

  const geoMock = {
    geocodeSearch: jasmine.createSpy('geocodeSearch'),
    hospitalsNear: jasmine.createSpy('hospitalsNear'),
    searchDoctors: jasmine.createSpy('searchDoctors'),
  };

  beforeEach(async () => {
    geoMock.geocodeSearch.calls.reset();
    geoMock.hospitalsNear.calls.reset();
    geoMock.searchDoctors.calls.reset();

    await TestBed.configureTestingModule({
      imports: [PatientFindCareComponent],
      providers: [{ provide: GeoBookingService, useValue: geoMock }],
    }).compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(PatientFindCareComponent);
    component = fixture.componentInstance;

    // Prevent Leaflet map initialization in unit tests.
    (component as any).initMap = jasmine.createSpy('initMap');
    (component as any).ensureMapReady = jasmine.createSpy('ensureMapReady');

    fixture.detectChanges();
  });

  it('shows an error if locationQuery is empty on search', () => {
    component.locationQuery = '';
    component.runGeocodeSearch();
    expect(component.error).toBe('Enter a place or address to search.');
  });

  it('runs geocodeSearch and updates lat/lng + label for a single hit', () => {
    geoMock.geocodeSearch.and.returnValue(
      of({
        results: [
          { lat: 29.6516, lng: -82.3248, display_name: 'Gainesville, FL' },
        ],
      })
    );

    component.locationQuery = 'Gainesville, FL';
    component.runGeocodeSearch();

    expect(geoMock.geocodeSearch).toHaveBeenCalledWith('Gainesville, FL', 5);
    expect(component.lat).toBe(29.6516);
    expect(component.lng).toBe(-82.3248);
    expect(component.selectedLocationLabel).toBe('Gainesville, FL');
  });

  it('pickGeocodeSuggestion updates locationQuery and search point', () => {
    const hit = { lat: 12.34, lng: 56.78, display_name: 'Someplace, City' };

    // updateSearchPoint calls ensureMapReady (spied), so we just validate state.
    component.pickGeocodeSuggestion(hit as any);

    expect(component.locationQuery).toContain('Someplace');
    expect(component.lat).toBe(12.34);
    expect(component.lng).toBe(56.78);
    expect(component.selectedLocationLabel).toBe('Someplace, City');
  });

  it('sets permission denied error when geolocation fails', () => {
    const mockGeolocation = {
      getCurrentPosition: (_success: any, error: any) => {
        error({ code: 1, PERMISSION_DENIED: 1 });
      },
    };

    // Karma/Chrome exposes navigator.geolocation via a getter; override via spy.
    spyOnProperty(navigator as any, 'geolocation', 'get').and.returnValue(mockGeolocation as any);

    component.useMyLocation();
    expect(component.error).toBe(
      'Location permission denied. Enable location or search manually.'
    );
  });

  it('loadHospitals calls API with current lat/lng/radius and sets hospitals', () => {
    (component as any).plotHospitals = jasmine.createSpy('plotHospitals');

    component.lat = 29.6516;
    component.lng = -82.3248;
    component.radiusKm = 25;

    geoMock.hospitalsNear.and.returnValue(
      of({
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
      })
    );

    component.loadHospitals();

    expect(geoMock.hospitalsNear).toHaveBeenCalledWith(29.6516, -82.3248, 25);
    expect(component.hospitals.length).toBe(1);
    expect(component.hospitals[0].name).toBe('UF Health Shands Hospital');
    expect((component as any).plotHospitals).toHaveBeenCalled();
  });

  it('loadDoctors sets doctors from API response', () => {
    component.lat = 10;
    component.lng = 20;
    component.radiusKm = 50;
    component.specialization = 'Internal Medicine';

    geoMock.searchDoctors.and.returnValue(
      of({
        doctors: [
          {
            id: '22222222-2222-2222-2222-222222222222',
            email: 'doc@healthonyx.demo',
            specialization: 'Internal Medicine',
            distance_km: 2.34,
            hospital_id: undefined,
          },
        ],
      })
    );

    component.loadDoctors();

    expect(geoMock.searchDoctors).toHaveBeenCalledWith(10, 20, 50, 'Internal Medicine');
    expect(component.doctors.length).toBe(1);
    expect(component.doctors[0].email).toBe('doc@healthonyx.demo');
  });
});


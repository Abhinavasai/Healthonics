import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { DoctorAvailabilityComponent } from './doctor-availability.component';
import { GeoBookingService } from '../../services/geo-booking.service';
import { AuthService } from '../../services/auth.service';

describe('DoctorAvailabilityComponent', () => {
  let fixture: ComponentFixture<DoctorAvailabilityComponent>;
  let component: DoctorAvailabilityComponent;

  const geoMock = {
    createSlot: jasmine.createSpy('createSlot'),
    listDoctorSlots: jasmine.createSpy('listDoctorSlots'),
    deleteSlot: jasmine.createSpy('deleteSlot')
  };

  const authMock = {
    getUser: jasmine.createSpy('getUser')
  };

  beforeEach(async () => {
    geoMock.createSlot.calls.reset();
    geoMock.listDoctorSlots.calls.reset();
    geoMock.deleteSlot.calls.reset();
    authMock.getUser.calls.reset();
    authMock.getUser.and.returnValue({ id: 'doctor-1' });

    await TestBed.configureTestingModule({
      imports: [DoctorAvailabilityComponent],
      providers: [
        { provide: GeoBookingService, useValue: geoMock },
        { provide: AuthService, useValue: authMock }
      ]
    }).compileComponents();
  });

  beforeEach(() => {
    geoMock.listDoctorSlots.and.returnValue(of({ slots: [] }));
    fixture = TestBed.createComponent(DoctorAvailabilityComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('loads doctor slots on init', () => {
    expect(geoMock.listDoctorSlots).toHaveBeenCalledWith('doctor-1');
  });

  it('shows validation error for invalid datetime input', () => {
    component.startAt = 'bad';
    component.endAt = 'bad';
    component.submit();
    expect(component.error).toContain('Use valid ISO-8601');
  });

  it('creates slot and refreshes list', () => {
    geoMock.createSlot.and.returnValue(of({ id: 'slot-1' }));
    geoMock.listDoctorSlots.and.returnValue(of({ slots: [{ id: 'slot-1', start_at: '2026-01-01T10:00:00Z', end_at: '2026-01-01T10:30:00Z', available: true }] }));
    component.startAt = '2026-01-01T10:00';
    component.endAt = '2026-01-01T10:30';

    component.submit();

    expect(geoMock.createSlot).toHaveBeenCalled();
    expect(component.message).toContain('Slot created');
  });

  it('deletes slot and refreshes list', () => {
    geoMock.deleteSlot.and.returnValue(of({ ok: true }));
    geoMock.listDoctorSlots.and.returnValue(of({ slots: [] }));

    component.remove('slot-1');

    expect(geoMock.deleteSlot).toHaveBeenCalledWith('slot-1');
    expect(component.message).toBe('Slot deleted.');
  });

  it('surfaces slot loading errors', () => {
    geoMock.listDoctorSlots.and.returnValue(throwError(() => ({ error: { error: 'boom' } })));
    component.loadSlots();
    expect(component.error).toBe('Could not load slots.');
  });
});

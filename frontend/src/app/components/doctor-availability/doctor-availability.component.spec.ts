import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { DoctorAvailabilityComponent } from './doctor-availability.component';
import { AuthService } from '../../services/auth.service';
import { GeoBookingService } from '../../services/geo-booking.service';

describe('DoctorAvailabilityComponent', () => {
  let fixture: ComponentFixture<DoctorAvailabilityComponent>;
  let component: DoctorAvailabilityComponent;
  let geoMock: jasmine.SpyObj<GeoBookingService>;

  beforeEach(async () => {
    geoMock = jasmine.createSpyObj<GeoBookingService>('GeoBookingService', [
      'listDoctorSlots',
      'createSlot',
      'deleteSlot'
    ]);

    geoMock.listDoctorSlots.and.returnValue(
      of({
        slots: [
          {
            id: 'slot-1',
            doctor_id: 'doc-1',
            start_at: '2026-04-29T14:00:00.000Z',
            end_at: '2026-04-29T14:30:00.000Z',
            available: true
          }
        ]
      })
    );
    geoMock.createSlot.and.returnValue(of({ id: 'slot-new' }));
    geoMock.deleteSlot.and.returnValue(of({ ok: true }));

    await TestBed.configureTestingModule({
      imports: [DoctorAvailabilityComponent],
      providers: [
        {
          provide: AuthService,
          useValue: {
            getUser: () => ({ id: 'doc-1', email: 'doc@healthonyx.demo', role: 'doctor' })
          }
        },
        { provide: GeoBookingService, useValue: geoMock }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(DoctorAvailabilityComponent);
    component = fixture.componentInstance;
    fixture.detectChanges(); // triggers ngOnInit
  });

  it('loads slots on init', () => {
    expect(geoMock.listDoctorSlots).toHaveBeenCalledWith('doc-1');
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('[data-cy="doctor-slot-row"]')).toBeTruthy();
  });

  it('creates a slot with UTC date+time payload', () => {
    component.selectedDate = '2026-04-30';
    component.onStartTimeChange('17:00'); // computes end time
    component.submit();

    const expectedStart = new Date(Date.UTC(2026, 3, 30, 17, 0, 0, 0)).toISOString();
    const expectedEnd = new Date(Date.UTC(2026, 3, 30, 17, 30, 0, 0)).toISOString();

    expect(geoMock.createSlot).toHaveBeenCalledWith({
      start_at: expectedStart,
      end_at: expectedEnd
    });
  });

  it('rejects invalid date/time', () => {
    component.selectedDate = '';
    component.startTime = '99:99';
    component.endTime = '99:99';

    component.submit();

    expect(geoMock.createSlot).not.toHaveBeenCalled();
    expect(component.error).toContain('valid UTC date and time');
  });

  it('deletes a slot and reloads list', () => {
    component.deleteSlot('slot-1');
    expect(geoMock.deleteSlot).toHaveBeenCalledWith('slot-1');
    expect(geoMock.listDoctorSlots).toHaveBeenCalledTimes(2); // init + reload after delete
  });
});


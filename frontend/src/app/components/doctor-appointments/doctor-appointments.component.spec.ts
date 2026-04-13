import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { DoctorAppointmentsComponent } from './doctor-appointments.component';
import { AppointmentsService } from '../../services/appointments.service';

describe('DoctorAppointmentsComponent', () => {
  let fixture: ComponentFixture<DoctorAppointmentsComponent>;
  let component: DoctorAppointmentsComponent;

  const appointmentsSvc = {
    listDoctor: jasmine.createSpy('listDoctor'),
    updateStatus: jasmine.createSpy('updateStatus')
  };

  beforeEach(async () => {
    appointmentsSvc.listDoctor.calls.reset();
    appointmentsSvc.updateStatus.calls.reset();
    appointmentsSvc.listDoctor.and.returnValue(
      of({
        appointments: [
          { id: '1', patient_id: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', doctor_id: 'doc', scheduled_at: '2026-01-01T10:00:00Z', reason: 'One', status: 'pending', created_at: '', updated_at: '' },
          { id: '2', patient_id: 'ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee', doctor_id: 'doc', scheduled_at: '2026-01-02T10:00:00Z', reason: 'Two', status: 'approved', created_at: '', updated_at: '' }
        ]
      })
    );
    appointmentsSvc.updateStatus.and.returnValue(of({}));

    await TestBed.configureTestingModule({
      imports: [DoctorAppointmentsComponent],
      providers: [
        { provide: AppointmentsService, useValue: appointmentsSvc },
        { provide: ActivatedRoute, useValue: { snapshot: { params: {} } } }
      ]
    }).compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(DoctorAppointmentsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('loads and computes status counts', () => {
    expect(component.appointments.length).toBe(2);
    expect(component.pendingCount).toBe(1);
    expect(component.approvedCount).toBe(1);
  });

  it('filters visible appointments by status', () => {
    component.activeFilter = 'pending';
    expect(component.visibleAppointments.length).toBe(1);
    expect(component.visibleAppointments[0].status).toBe('pending');
  });

  it('updates status and reloads list', () => {
    component.updateStatus('1', 'approved');
    expect(appointmentsSvc.updateStatus).toHaveBeenCalledWith('1', 'approved');
    expect(appointmentsSvc.listDoctor).toHaveBeenCalled();
  });
});


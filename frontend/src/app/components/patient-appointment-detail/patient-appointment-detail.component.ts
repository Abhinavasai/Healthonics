import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { Subscription, forkJoin, of } from 'rxjs';
import { catchError, distinctUntilChanged, filter, finalize, map, switchMap, tap } from 'rxjs/operators';
import {
  Appointment,
  AppointmentActivity,
  AppointmentsService
} from '../../services/appointments.service';
import { AppointmentActivityTimelineComponent } from '../appointment-activity-timeline/appointment-activity-timeline.component';

@Component({
  selector: 'app-patient-appointment-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe, TitleCasePipe, AppointmentActivityTimelineComponent],
  templateUrl: './patient-appointment-detail.component.html',
  styleUrl: './patient-appointment-detail.component.scss'
})
export class PatientAppointmentDetailComponent implements OnInit, OnDestroy {
  appointment: Appointment | null = null;
  activities: AppointmentActivity[] = [];
  loading = false;
  activityError = '';
  pageError = '';
  private sub?: Subscription;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private appointmentsService: AppointmentsService
  ) {}

  ngOnInit(): void {
    this.sub = this.route.paramMap
      .pipe(
        map((p) => p.get('id')),
        filter((id): id is string => !!id),
        distinctUntilChanged(),
        switchMap((id) => {
          this.loading = true;
          this.pageError = '';
          this.activityError = '';
          this.appointment = null;
          this.activities = [];
          return forkJoin({
            appointment: this.appointmentsService.getById(id),
            activities: this.appointmentsService.getActivity(id).pipe(
              catchError(() => {
                this.activityError = 'Unable to load activity history.';
                return of([] as AppointmentActivity[]);
              })
            )
          }).pipe(
            tap(({ appointment, activities }) => {
              this.appointment = appointment;
              this.activities = activities;
            }),
            catchError((err) => {
              this.pageError = err?.error?.error ?? 'Unable to load appointment';
              this.appointment = null;
              this.activities = [];
              return of(undefined);
            }),
            finalize(() => (this.loading = false))
          );
        })
      )
      .subscribe();
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  back(): void {
    void this.router.navigate(['/patient/appointments']);
  }
}

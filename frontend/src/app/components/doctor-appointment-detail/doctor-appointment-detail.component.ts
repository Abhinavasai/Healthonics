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
  selector: 'app-doctor-appointment-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, DatePipe, TitleCasePipe, AppointmentActivityTimelineComponent],
  templateUrl: './doctor-appointment-detail.component.html',
  styleUrl: './doctor-appointment-detail.component.scss'
})
export class DoctorAppointmentDetailComponent implements OnInit, OnDestroy {
  appointment: Appointment | null = null;
  activities: AppointmentActivity[] = [];
  loading = false;
  deciding = false;
  activityError = '';
  pageError = '';
  private appointmentId = '';
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
        tap((id) => (this.appointmentId = id)),
        switchMap((id) => this.loadBundle(id))
      )
      .subscribe();
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  private loadBundle(id: string) {
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
  }

  decide(status: 'approved' | 'rejected'): void {
    if (!this.appointmentId || this.deciding) {
      return;
    }
    this.deciding = true;
    this.pageError = '';
    this.appointmentsService
      .updateStatus(this.appointmentId, status)
      .pipe(
        switchMap(() =>
          forkJoin({
            appointment: this.appointmentsService.getById(this.appointmentId),
            activities: this.appointmentsService.getActivity(this.appointmentId).pipe(
              catchError(() => {
                this.activityError = 'Unable to load activity history.';
                return of([] as AppointmentActivity[]);
              })
            )
          })
        ),
        tap(({ appointment, activities }) => {
          this.appointment = appointment;
          this.activities = activities;
          this.activityError = '';
        }),
        catchError((err) => {
          this.pageError = err?.error?.error ?? 'Unable to update status';
          return of(undefined);
        }),
        finalize(() => (this.deciding = false))
      )
      .subscribe();
  }

  back(): void {
    void this.router.navigate(['/doctor/appointments']);
  }
}

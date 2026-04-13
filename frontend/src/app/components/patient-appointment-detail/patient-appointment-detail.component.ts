import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule, DatePipe, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { Subscription, forkJoin, of } from 'rxjs';
import { catchError, distinctUntilChanged, filter, finalize, map, switchMap, tap } from 'rxjs/operators';
import {
  Appointment,
  AppointmentActivity,
  AppointmentStatus,
  AppointmentsService
} from '../../services/appointments.service';
import { AppointmentActivityTimelineComponent } from '../appointment-activity-timeline/appointment-activity-timeline.component';

@Component({
  selector: 'app-patient-appointment-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, DatePipe, TitleCasePipe, AppointmentActivityTimelineComponent],
  templateUrl: './patient-appointment-detail.component.html',
  styleUrl: './patient-appointment-detail.component.scss'
})
export class PatientAppointmentDetailComponent implements OnInit, OnDestroy {
  appointment: Appointment | null = null;
  activities: AppointmentActivity[] = [];
  loading = false;
  activityRefreshing = false;
  actionLoading = false;
  cancelReason = '';
  rescheduleReason = '';
  /** Local datetime string for `<input type="datetime-local">` */
  rescheduleAtLocal = '';
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
        switchMap((id) => {
          this.appointmentId = id;
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

  refreshActivities(): void {
    if (!this.appointmentId) {
      return;
    }
    this.activityRefreshing = true;
    this.activityError = '';
    this.appointmentsService
      .getActivity(this.appointmentId)
      .pipe(
        finalize(() => (this.activityRefreshing = false)),
        catchError(() => {
          this.activityError = 'Unable to load activity history.';
          return of([] as AppointmentActivity[]);
        })
      )
      .subscribe((rows) => (this.activities = rows));
  }

  back(): void {
    void this.router.navigate(['/patient/appointments']);
  }

  statusLabel(s: AppointmentStatus): string {
    return s.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
  }

  private isFuture(a: Appointment): boolean {
    return new Date(a.scheduled_at).getTime() > Date.now();
  }

  canCancel(a: Appointment): boolean {
    return (
      (a.status === 'pending' || a.status === 'approved' || a.status === 'reschedule_requested') &&
      this.isFuture(a)
    );
  }

  canRequestReschedule(a: Appointment): boolean {
    return (a.status === 'pending' || a.status === 'approved') && this.isFuture(a);
  }

  cancel(): void {
    if (!this.appointmentId || !this.appointment || this.actionLoading || !this.canCancel(this.appointment)) {
      return;
    }
    this.actionLoading = true;
    this.pageError = '';
    this.appointmentsService
      .cancel(this.appointmentId, this.cancelReason.trim() || undefined)
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
          this.cancelReason = '';
        }),
        catchError((err) => {
          this.pageError = err?.error?.error ?? 'Unable to cancel';
          return of(undefined);
        }),
        finalize(() => (this.actionLoading = false))
      )
      .subscribe();
  }

  requestReschedule(): void {
    if (!this.appointmentId || !this.appointment || this.actionLoading || !this.canRequestReschedule(this.appointment)) {
      return;
    }
    const raw = this.rescheduleAtLocal?.trim();
    if (!raw) {
      this.pageError = 'Choose a new date and time for your request.';
      return;
    }
    const iso = new Date(raw).toISOString();
    this.actionLoading = true;
    this.pageError = '';
    this.appointmentsService
      .requestReschedule(this.appointmentId, iso, this.rescheduleReason.trim() || undefined)
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
          this.rescheduleAtLocal = '';
          this.rescheduleReason = '';
        }),
        catchError((err) => {
          this.pageError = err?.error?.error ?? 'Unable to submit reschedule request';
          return of(undefined);
        }),
        finalize(() => (this.actionLoading = false))
      )
      .subscribe();
  }
}

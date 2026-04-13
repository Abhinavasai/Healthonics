import { Routes } from '@angular/router';
import { RegisterComponent } from './components/register/register.component';
import { LoginComponent } from './components/login/login.component';
import { LandingComponent } from './components/landing/landing.component';
import { AppShellComponent } from './components/app-shell/app-shell.component';
import { DashboardPlaceholderComponent } from './components/dashboard-placeholder/dashboard-placeholder.component';
import { AdminShellComponent } from './components/admin-shell/admin-shell.component';
import { AdminAuditComponent } from './components/admin-audit/admin-audit.component';
import { PatientAppointmentsComponent } from './components/patient-appointments/patient-appointments.component';
import { PatientAppointmentDetailComponent } from './components/patient-appointment-detail/patient-appointment-detail.component';
import { DoctorAppointmentsComponent } from './components/doctor-appointments/doctor-appointments.component';
import { DoctorAppointmentDetailComponent } from './components/doctor-appointment-detail/doctor-appointment-detail.component';
import { PatientFindCareComponent } from './components/patient-find-care/patient-find-care.component';
import { DoctorAvailabilityComponent } from './components/doctor-availability/doctor-availability.component';
import { MessagesComponent } from './components/messages/messages.component';
import { authGuard } from './guards/auth.guard';
import { roleGuard } from './guards/role.guard';

export const routes: Routes = [
  { path: '', component: LandingComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  {
    path: '',
    component: AppShellComponent,
    canActivate: [authGuard],
    children: [
      {
        path: 'patient',
        data: { roles: ['patient'] },
        canActivate: [roleGuard],
        children: [
          { path: '', pathMatch: 'full', redirectTo: 'appointments' },
          {
            path: 'find-care',
            component: PatientFindCareComponent,
            data: { title: 'Find care', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments/:id',
            component: PatientAppointmentDetailComponent,
            data: { title: 'Appointment details', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments',
            component: PatientAppointmentsComponent,
            data: { title: 'Patient Appointments', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages',
            component: MessagesComponent,
            data: { title: 'Messages', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages/:threadId',
            component: MessagesComponent,
            data: { title: 'Messages', roles: ['patient'] },
            canActivate: [roleGuard]
          }
        ]
      },
      {
        path: 'doctor',
        data: { roles: ['doctor'] },
        canActivate: [roleGuard],
        children: [
          { path: '', pathMatch: 'full', redirectTo: 'appointments' },
          {
            path: 'availability',
            component: DoctorAvailabilityComponent,
            data: { title: 'Availability', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments/:id',
            component: DoctorAppointmentDetailComponent,
            data: { title: 'Request details', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments',
            component: DoctorAppointmentsComponent,
            data: { title: 'Doctor Appointments', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages',
            component: MessagesComponent,
            data: { title: 'Messages', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages/:threadId',
            component: MessagesComponent,
            data: { title: 'Messages', roles: ['doctor'] },
            canActivate: [roleGuard]
          }
        ]
      },
      {
        path: 'admin',
        component: AdminShellComponent,
        data: { roles: ['admin'] },
        canActivate: [roleGuard],
        children: [
          {
            path: '',
            component: DashboardPlaceholderComponent,
            data: { title: 'Admin', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'audit',
            component: AdminAuditComponent,
            data: { title: 'Audit log', roles: ['admin'] },
            canActivate: [roleGuard]
          }
        ]
      }
    ]
  },
  { path: '**', redirectTo: '/' }
];

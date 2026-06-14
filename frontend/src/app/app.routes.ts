import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';
import { roleGuard } from './guards/role.guard';

// All feature routes use loadComponent for lazy loading — each route becomes its own
// JS chunk fetched on demand, reducing the initial bundle from 3.8 MB to ~300 KB.
export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/landing/landing.component').then(m => m.LandingComponent)
  },
  {
    path: 'register',
    loadComponent: () => import('./components/register/register.component').then(m => m.RegisterComponent)
  },
  {
    path: 'features/:slug',
    loadComponent: () => import('./components/feature-detail/feature-detail.component').then(m => m.FeatureDetailComponent)
  },
  {
    path: 'login',
    loadComponent: () => import('./components/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: '',
    loadComponent: () => import('./components/app-shell/app-shell.component').then(m => m.AppShellComponent),
    canActivate: [authGuard],
    children: [
      {
        path: 'patient',
        data: { roles: ['patient'] },
        canActivate: [roleGuard],
        children: [
          { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
          {
            path: 'dashboard',
            loadComponent: () => import('./components/patient-dashboard/patient-dashboard.component').then(m => m.PatientDashboardComponent),
            data: { title: 'Dashboard', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'find-care',
            loadComponent: () => import('./components/patient-find-care/patient-find-care.component').then(m => m.PatientFindCareComponent),
            data: { title: 'Find care', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments/:id',
            loadComponent: () => import('./components/patient-appointment-detail/patient-appointment-detail.component').then(m => m.PatientAppointmentDetailComponent),
            data: { title: 'Appointment details', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments',
            loadComponent: () => import('./components/patient-appointments/patient-appointments.component').then(m => m.PatientAppointmentsComponent),
            data: { title: 'Patient Appointments', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'documents',
            loadComponent: () => import('./components/patient-documents/patient-documents.component').then(m => m.PatientDocumentsComponent),
            data: { title: 'My documents', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages',
            loadComponent: () => import('./components/messages/messages.component').then(m => m.MessagesComponent),
            data: { title: 'Messages', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages/:threadId',
            loadComponent: () => import('./components/messages/messages.component').then(m => m.MessagesComponent),
            data: { title: 'Messages', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'files',
            loadComponent: () => import('./components/patient-files/patient-files.component').then(m => m.PatientFilesComponent),
            data: { title: 'My files', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          { path: 'my-files', pathMatch: 'full', redirectTo: 'files' },
          {
            path: 'prescriptions',
            loadComponent: () => import('./components/patient-prescriptions/patient-prescriptions.component').then(m => m.PatientPrescriptionsComponent),
            data: { title: 'My prescriptions', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'notifications',
            loadComponent: () => import('./components/notifications-inbox/notifications-inbox.component').then(m => m.NotificationsInboxComponent),
            data: { title: 'Notifications', roles: ['patient'] },
            canActivate: [roleGuard]
          },
          {
            path: 'symptom-check',
            loadComponent: () => import('./components/symptom-checker/symptom-checker.component').then(m => m.SymptomCheckerComponent),
            data: { title: 'Symptom Checker', roles: ['patient'] },
            canActivate: [roleGuard]
          }
        ]
      },
      {
        path: 'doctor',
        data: { roles: ['doctor'] },
        canActivate: [roleGuard],
        children: [
          { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
          {
            path: 'dashboard',
            loadComponent: () => import('./components/doctor-dashboard/doctor-dashboard.component').then(m => m.DoctorDashboardComponent),
            data: { title: 'Dashboard', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'availability',
            loadComponent: () => import('./components/doctor-availability/doctor-availability.component').then(m => m.DoctorAvailabilityComponent),
            data: { title: 'Availability', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments/:id',
            loadComponent: () => import('./components/doctor-appointment-detail/doctor-appointment-detail.component').then(m => m.DoctorAppointmentDetailComponent),
            data: { title: 'Request details', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'appointments',
            loadComponent: () => import('./components/doctor-appointments/doctor-appointments.component').then(m => m.DoctorAppointmentsComponent),
            data: { title: 'Doctor Appointments', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'documents',
            loadComponent: () => import('./components/doctor-documents-list/doctor-documents-list.component').then(m => m.DoctorDocumentsListComponent),
            data: { title: 'Patient documents', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'documents/:documentId',
            loadComponent: () => import('./components/doctor-document-detail/doctor-document-detail.component').then(m => m.DoctorDocumentDetailComponent),
            data: { title: 'Document', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages',
            loadComponent: () => import('./components/messages/messages.component').then(m => m.MessagesComponent),
            data: { title: 'Messages', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'messages/:threadId',
            loadComponent: () => import('./components/messages/messages.component').then(m => m.MessagesComponent),
            data: { title: 'Messages', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'notifications',
            loadComponent: () => import('./components/notifications-inbox/notifications-inbox.component').then(m => m.NotificationsInboxComponent),
            data: { title: 'Notifications', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'second-opinions',
            loadComponent: () => import('./components/second-opinion-board/second-opinion-board.component').then(m => m.SecondOpinionBoardComponent),
            data: { title: 'Second Opinions', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'patients',
            loadComponent: () => import('./components/doctor-patient-panel/doctor-patient-panel.component').then(m => m.DoctorPatientPanelComponent),
            data: { title: 'My Patients', roles: ['doctor'] },
            canActivate: [roleGuard]
          },
          {
            path: 'waiting-room',
            loadComponent: () => import('./components/waiting-room/waiting-room.component').then(m => m.WaitingRoomComponent),
            data: { title: 'Waiting Room', roles: ['doctor'] },
            canActivate: [roleGuard]
          }
        ]
      },
      {
        path: 'admin',
        loadComponent: () => import('./components/admin-shell/admin-shell.component').then(m => m.AdminShellComponent),
        data: { roles: ['admin'] },
        canActivate: [roleGuard],
        children: [
          {
            path: '',
            loadComponent: () => import('./components/dashboard-placeholder/dashboard-placeholder.component').then(m => m.DashboardPlaceholderComponent),
            data: { title: 'Admin', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'audit',
            loadComponent: () => import('./components/admin-audit/admin-audit.component').then(m => m.AdminAuditComponent),
            data: { title: 'Audit log', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'notifications',
            loadComponent: () => import('./components/admin-notifications-log/admin-notifications-log.component').then(m => m.AdminNotificationsLogComponent),
            data: { title: 'Notifications', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'knowledge',
            loadComponent: () => import('./components/admin-knowledge/admin-knowledge.component').then(m => m.AdminKnowledgeComponent),
            data: { title: 'Knowledge', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'management',
            loadComponent: () => import('./components/admin-management/admin-management.component').then(m => m.AdminManagementComponent),
            data: { title: 'Management', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'ai',
            loadComponent: () => import('./components/admin-ai-observability/admin-ai-observability.component').then(m => m.AdminAiObservabilityComponent),
            data: { title: 'AI controls', roles: ['admin'] },
            canActivate: [roleGuard]
          },
          {
            path: 'slo',
            loadComponent: () => import('./components/admin-slo-dashboard/admin-slo-dashboard.component').then(m => m.AdminSloDashboardComponent),
            data: { title: 'SLO Dashboard', roles: ['admin'] },
            canActivate: [roleGuard]
          }
        ]
      }
    ]
  },
  { path: '**', redirectTo: '/' }
];

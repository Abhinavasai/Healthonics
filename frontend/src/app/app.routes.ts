import { Routes } from '@angular/router';
import { RegisterComponent } from './components/register/register.component';
import { LoginComponent } from './components/login/login.component';
import { DashboardPlaceholderComponent } from './components/dashboard-placeholder/dashboard-placeholder.component';
import { authGuard } from './guards/auth.guard';
import { roleGuard } from './guards/role.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/register', pathMatch: 'full' },
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  {
    path: 'patient',
    component: DashboardPlaceholderComponent,
    data: { title: 'Patient', roles: ['patient'] },
    canActivate: [authGuard, roleGuard]
  },
  {
    path: 'doctor',
    component: DashboardPlaceholderComponent,
    data: { title: 'Doctor', roles: ['doctor'] },
    canActivate: [authGuard, roleGuard]
  },
  {
    path: 'admin',
    component: DashboardPlaceholderComponent,
    data: { title: 'Admin', roles: ['admin'] },
    canActivate: [authGuard, roleGuard]
  },
  { path: '**', redirectTo: '/register' }
];

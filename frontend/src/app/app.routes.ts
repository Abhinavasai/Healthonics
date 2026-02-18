import { Routes } from '@angular/router';
import { RegisterComponent } from './components/register/register.component';
import { LoginComponent } from './components/login/login.component';
import { DashboardPlaceholderComponent } from './components/dashboard-placeholder/dashboard-placeholder.component';

export const routes: Routes = [
  { path: '', redirectTo: '/register', pathMatch: 'full' },
  { path: 'register', component: RegisterComponent },
  { path: 'login', component: LoginComponent },
  { path: 'patient', component: DashboardPlaceholderComponent, data: { title: 'Patient' } },
  { path: 'doctor', component: DashboardPlaceholderComponent, data: { title: 'Doctor' } },
  { path: 'admin', component: DashboardPlaceholderComponent, data: { title: 'Admin' } },
  { path: '**', redirectTo: '/register' }
];

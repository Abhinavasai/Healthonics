import { inject } from '@angular/core';
import { Router, CanActivateFn, ActivatedRouteSnapshot } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const roleGuard: CanActivateFn = (route: ActivatedRouteSnapshot) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const user = auth.getUser();

  if (!user) {
    router.navigate(['/login']);
    return false;
  }

  const allowedRoles = route.data['roles'] as string[] | undefined;
  if (!allowedRoles || allowedRoles.length === 0) {
    console.error('roleGuard: route is missing data.roles — access denied by default', route);
    return false;
  }

  if (allowedRoles.includes(user.role)) {
    return true;
  }

  // Redirect to role-specific dashboard when wrong role
  const path = user.role === 'admin' ? '/admin' : user.role === 'doctor' ? '/doctor' : '/patient';
  router.navigate([path]);
  return false;
};

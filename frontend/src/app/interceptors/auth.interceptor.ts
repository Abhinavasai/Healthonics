import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const token = auth.getToken();

  if (token && req.url.startsWith('/api')) {
    req = req.clone({
      setHeaders: { Authorization: `Bearer ${token}` }
    });
  }

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      // Avoid global auto-logout on any 401: many endpoints can legitimately return 401/403
      // (e.g. optional polling) and should be handled by the calling UI.
      // We only hard-redirect when the profile endpoint rejects the token.
      if (err.status === 401 && req.url.startsWith('/api/me')) {
        auth.logout();
        router.navigate(['/login']);
      }
      return throwError(() => err);
    })
  );
};

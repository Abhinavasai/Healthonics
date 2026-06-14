import { HttpInterceptorFn, HttpErrorResponse, HttpRequest, HttpHandlerFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, switchMap, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return next(attachToken(req, auth)).pipe(
    catchError((err: HttpErrorResponse) => {
      // Skip refresh loop for the refresh and logout endpoints themselves.
      if (err.status === 401 && !req.url.includes('/api/auth/')) {
        const rt = auth.getRefreshToken();
        if (rt) {
          return auth.refreshAccessToken().pipe(
            switchMap((res) => {
              // Retry original request with the new token.
              return next(attachToken(req, auth));
            }),
            catchError(() => {
              // Refresh failed — force logout.
              auth.logout();
              router.navigate(['/login']);
              return throwError(() => err);
            })
          );
        }
        auth.logout();
        router.navigate(['/login']);
      }
      return throwError(() => err);
    })
  );
};

function attachToken(req: HttpRequest<unknown>, auth: AuthService): HttpRequest<unknown> {
  const token = auth.getToken();
  if (token && req.url.startsWith('/api')) {
    return req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return req;
}

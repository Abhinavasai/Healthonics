import { ErrorHandler, Injectable } from '@angular/core';

@Injectable()
export class GlobalErrorHandler implements ErrorHandler {
  handleError(error: unknown): void {
    // Log unhandled errors to the console; swap for Sentry/DataDog in production.
    console.error('[GlobalErrorHandler]', error);
  }
}

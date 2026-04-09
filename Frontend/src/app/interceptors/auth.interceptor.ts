import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, switchMap } from 'rxjs/operators';
import { AuthService } from '../services/auth.service';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private authService: AuthService) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const isAuthRoute = req.url.includes('/auth/login') || req.url.includes('/auth/register') || req.url.includes('/auth/refresh');
    const token = this.authService.getAccessToken();

    if (token) {
      req = req.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`
        }
      });
    }

    return next.handle(req).pipe(
      catchError(error => {
        if (error.status === 401 && !isAuthRoute) {
          return this.authService.refreshAccessToken().pipe(
            switchMap(newToken => {
              if (!newToken) {
                this.authService.logoutLocalOnly();
                return throwError(() => error);
              }

              const retryReq = req.clone({
                setHeaders: {
                  Authorization: `Bearer ${newToken}`
                }
              });
              return next.handle(retryReq);
            }),
            catchError(refreshErr => {
              this.authService.logoutLocalOnly();
              return throwError(() => refreshErr);
            })
          );
        }
        return throwError(() => error);
      })
    );
  }
}

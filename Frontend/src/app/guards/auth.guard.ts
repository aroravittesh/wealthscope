import { Injectable } from '@angular/core';
import { Router, CanActivate, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { AuthService } from '../services/auth.service';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<boolean> {
    if (this.authService.isAuthenticated()) {
      return this.authService.isAuthenticated$.pipe(
        map(isAuth => {
          if (!isAuth) {
            this.router.navigate(['/auth/login']);
          }
          return isAuth;
        })
      );
    }

    this.router.navigate(['/auth/login'], { queryParams: { returnUrl: state.url } });
    return new Observable(observer => {
      observer.next(false);
      observer.complete();
    });
  }
}

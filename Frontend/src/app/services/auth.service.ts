import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, map, switchMap, tap } from 'rxjs/operators';
import { User } from '../models';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  private currentUserSubject = new BehaviorSubject<User | null>(this.getUserFromStorage());
  public currentUser$ = this.currentUserSubject.asObservable();

  private isAuthenticatedSubject = new BehaviorSubject<boolean>(this.hasToken());
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {
    this.bootstrapSession();
    this.checkTokenExpiry();
  }

  // ========================
  // REGISTER (Backend)
  // ========================
  register(email: string, password: string, riskPreference: string): Observable<{ message: string }> {
    const payload = {
      email,
      password,
      risk_preference: riskPreference.toUpperCase()
    };

    return this.http.post<{ message: string }>(`${this.apiUrl}/auth/register`, payload);
  }

  // ========================
  // LOGIN (Backend)
  // ========================
  login(email: string, password: string): Observable<void> {
    const payload = { email, password };

    return this.http.post<{ access_token: string; refresh_token: string }>(
      `${this.apiUrl}/auth/login`,
      payload
    ).pipe(
      tap(response => {
        localStorage.setItem('authToken', response.access_token);
        localStorage.setItem('refreshToken', response.refresh_token);
        this.isAuthenticatedSubject.next(true);
      }),
      switchMap(() => this.getProfile()),
      tap(profile => this.setUserFromProfile(profile)),
      map(() => undefined)
    );
  }

  // ========================
  // PROFILE
  // ========================

  getProfile(): Observable<{ email: string; risk_preference: string; role?: 'USER' | 'ADMIN' }> {
    return this.http.get<{ email: string; risk_preference: string; role?: 'USER' | 'ADMIN' }>(
      `${this.apiUrl}/auth/profile`
    );
  }

  updateRiskPreference(riskPreference: string): Observable<{ email: string; risk_preference: string }> {
    const payload = { risk_preference: riskPreference.toUpperCase() };
    return this.http.put<{ email: string; risk_preference: string }>(`${this.apiUrl}/auth/profile`, payload);
  }

  // ========================
  // CHANGE PASSWORD
  // ========================

  changePassword(oldPassword: string, newPassword: string): Observable<{ message: string }> {
    const payload = {
      old_password: oldPassword,
      new_password: newPassword
    };
    return this.http.post<{ message: string }>(`${this.apiUrl}/auth/change-password`, payload);
  }

  // ========================
  // LOGOUT
  // ========================
  logout(): void {
    const refreshToken = localStorage.getItem('refreshToken');
    if (refreshToken) {
      this.http.post<{ message: string }>(
        `${this.apiUrl}/auth/logout`,
        { refresh_token: refreshToken }
      ).pipe(catchError(() => of(null))).subscribe();
    }
    this.clearSession();
  }

  logoutLocalOnly(): void {
    this.clearSession();
  }

  refreshAccessToken(): Observable<string> {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      return of('');
    }

    return this.http.post<{ access_token: string; refresh_token: string }>(
      `${this.apiUrl}/auth/refresh`,
      { refresh_token: refreshToken }
    ).pipe(
      tap(tokens => {
        localStorage.setItem('authToken', tokens.access_token);
        localStorage.setItem('refreshToken', tokens.refresh_token);
        this.isAuthenticatedSubject.next(true);
      }),
      map(tokens => tokens.access_token),
      catchError(() => {
        this.clearSession();
        return of('');
      })
    );
  }

  getAccessToken(): string | null {
    return localStorage.getItem('authToken');
  }

  getRefreshToken(): string | null {
    return localStorage.getItem('refreshToken');
  }

  private clearSession(): void {
    localStorage.removeItem('authToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);
  }

  // ========================
  // HELPERS
  // ========================
  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  isAuthenticated(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  private hasToken(): boolean {
    const token = localStorage.getItem('authToken');
    return !!token && !this.isTokenExpired();
  }

  private getUserFromStorage(): User | null {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  }

  private checkTokenExpiry(): void {
    setInterval(() => {
      if (this.hasToken() && this.isTokenExpired()) {
        this.logoutLocalOnly();
      }
    }, 60000);
  }

  private isTokenExpired(): boolean {
    const token = localStorage.getItem('authToken');
    if (!token) return true;

    try {
      const decoded = JSON.parse(atob(token.split('.')[1]));
      return decoded.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }

  private setUserFromProfile(profile: { email: string; risk_preference: string; role?: 'USER' | 'ADMIN' }): void {
    const roleFromToken = this.getRoleFromToken();
    const user: User = {
      email: profile.email,
      riskPreference: profile.risk_preference,
      role: profile.role ?? roleFromToken
    };
    localStorage.setItem('user', JSON.stringify(user));
    this.currentUserSubject.next(user);
  }

  private getRoleFromToken(): 'USER' | 'ADMIN' | undefined {
    const token = localStorage.getItem('authToken');
    if (!token) return undefined;

    try {
      const decoded = JSON.parse(atob(token.split('.')[1]));
      const role = decoded?.role;
      if (role === 'ADMIN' || role === 'USER') {
        return role;
      }
      return undefined;
    } catch {
      return undefined;
    }
  }

  private bootstrapSession(): void {
    if (!localStorage.getItem('authToken')) {
      return;
    }

    if (this.isTokenExpired()) {
      this.clearSession();
      return;
    }

    this.isAuthenticatedSubject.next(true);

    if (!this.getCurrentUser()) {
      this.getProfile().pipe(
        tap(profile => this.setUserFromProfile(profile)),
        catchError(() => {
          this.clearSession();
          return of(null);
        })
      ).subscribe();
    }
  }
}
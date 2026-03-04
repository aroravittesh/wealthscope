import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { User, AuthResponse } from '../models';
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
      tap(() => {
        // Optional: fetch profile later to populate currentUser if needed
      }),
    ) as unknown as Observable<void>;
  }

  // ========================
  // PROFILE
  // ========================

  getProfile(): Observable<{ email: string; risk_preference: string }> {
    return this.http.get<{ email: string; risk_preference: string }>(`${this.apiUrl}/auth/profile`);
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

  private handleAuthResponse(response: AuthResponse): void {
    localStorage.setItem('authToken', response.token);
    localStorage.setItem('refreshToken', response.refreshToken);
    localStorage.setItem('user', JSON.stringify(response.user));
    this.currentUserSubject.next(response.user);
    this.isAuthenticatedSubject.next(true);
  }

  private hasToken(): boolean {
    return !!localStorage.getItem('authToken');
  }

  private getUserFromStorage(): User | null {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  }

  private checkTokenExpiry(): void {
    setInterval(() => {
      if (this.hasToken() && this.isTokenExpired()) {
        this.logout();
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
}
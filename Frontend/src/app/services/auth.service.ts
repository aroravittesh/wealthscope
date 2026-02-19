import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, of, throwError } from 'rxjs';
import { tap, delay } from 'rxjs/operators';
import { User, AuthResponse } from '../models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  private currentUserSubject = new BehaviorSubject<User | null>(this.getUserFromStorage());
  public currentUser$ = this.currentUserSubject.asObservable();

  private isAuthenticatedSubject = new BehaviorSubject<boolean>(this.hasToken());
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  // Mock users database
  private mockUsers: Map<string, { password: string; user: User }> = new Map([
    ['test@example.com', {
      password: 'password123',
      user: {
        id: '1',
        email: 'test@example.com',
        fullName: 'Test User',
        role: 'USER',
        createdAt: new Date(),
        updatedAt: new Date()
      }
    }],
    ['demo@finsight.com', {
      password: 'Demo@123',
      user: {
        id: '2',
        email: 'demo@finsight.com',
        fullName: 'Demo Account',
        role: 'USER',
        createdAt: new Date(),
        updatedAt: new Date()
      }
    }]
  ]);

  constructor() {
    this.checkTokenExpiry();
  }

  // ========================
  // REGISTER (Mock Only)
  // ========================
  register(email: string, password: string, fullName: string): Observable<AuthResponse> {

    if (this.mockUsers.has(email)) {
      return throwError(() => ({
        error: { message: 'Email already registered' }
      })).pipe(delay(800));
    }

    const newUser: User = {
      id: Math.random().toString(36).substring(2, 9),
      email,
      fullName,
      role: 'USER',
      createdAt: new Date(),
      updatedAt: new Date()
    };

    this.mockUsers.set(email, {
      password,
      user: newUser
    });

    const mockResponse: AuthResponse = {
      token: this.generateMockToken(newUser),
      refreshToken: this.generateMockToken(newUser),
      user: newUser
    };

    return of(mockResponse).pipe(
      delay(1000),
      tap(response => this.handleAuthResponse(response))
    );
  }

  // ========================
  // LOGIN (Mock Only)
  // ========================
  login(email: string, password: string): Observable<AuthResponse> {

    const mockUser = this.mockUsers.get(email);

    if (!mockUser || mockUser.password !== password) {
      return throwError(() => ({
        error: { message: 'Invalid email or password' }
      })).pipe(delay(800));
    }

    const mockResponse: AuthResponse = {
      token: this.generateMockToken(mockUser.user),
      refreshToken: this.generateMockToken(mockUser.user),
      user: mockUser.user
    };

    return of(mockResponse).pipe(
      delay(800),
      tap(response => this.handleAuthResponse(response))
    );
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

  private generateMockToken(user: User): string {
    const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
    const payload = btoa(JSON.stringify({
      sub: user.id,
      email: user.email,
      name: user.fullName,
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
    }));
    const signature = btoa('mock-signature');
    return `${header}.${payload}.${signature}`;
  }
}
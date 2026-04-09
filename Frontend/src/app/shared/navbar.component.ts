import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { User } from '../models';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  template: `
    <nav class="bg-slate-800 bg-opacity-50 backdrop-filter backdrop-blur-lg border-b border-slate-700 sticky top-0 z-50 transition-all duration-300 hover:bg-opacity-70">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center gap-3 cursor-pointer group" [routerLink]="isAuthenticated ? '/dashboard' : '/'">
            <div class="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center transform group-hover:scale-110 group-hover:shadow-lg group-hover:shadow-blue-500/50 transition-all duration-300">
              <span class="text-lg font-bold text-white">W</span>
            </div>
            <span class="text-white font-bold text-xl hidden sm:block group-hover:text-blue-300 transition-colors">WealthScope</span>
          </div>

          <div *ngIf="isAuthenticated; else publicNav" class="hidden md:flex items-center gap-8">
            <a routerLink="/dashboard" routerLinkActive="text-blue-400" [routerLinkActiveOptions]="{ exact: true }" class="text-slate-300 hover:text-white transition-all duration-300">Dashboard</a>
            <a routerLink="/portfolio" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Portfolios</a>
            <a routerLink="/analytics" class="text-slate-300 hover:text-white transition-all duration-300">Analytics</a>
            <a class="text-slate-500 cursor-not-allowed">Watchlist</a>
          </div>
          <ng-template #publicNav></ng-template>

          <div *ngIf="isAuthenticated; else guestActions" class="flex items-center gap-3">
            <a routerLink="/profile" class="text-slate-300 hover:text-white text-sm">Profile</a>
            <button type="button" (click)="onLogout()" class="bg-red-600 hover:bg-red-700 text-white text-sm font-semibold px-4 py-2 rounded-lg">
              Logout
            </button>
            <div *ngIf="currentUser" class="hidden md:block">
              <select [(ngModel)]="currentUser.role" (change)="onRoleChange()" class="bg-slate-700 text-xs text-slate-300 rounded px-2 py-1 border border-slate-600">
                <option value="USER">USER</option>
                <option value="ADMIN">ADMIN</option>
              </select>
            </div>
          </div>
          <ng-template #guestActions>
            <div class="flex items-center gap-3">
              <a routerLink="/auth/login" class="text-slate-300 hover:text-white font-semibold">Login</a>
              <a routerLink="/auth/signup" class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-semibold px-4 py-2 rounded-lg transition">
                Sign Up
              </a>
            </div>
          </ng-template>

        </div>
      </div>
    </nav>
  `,
  styles: []
})
export class NavbarComponent implements OnInit {
  currentUser: User | null = null;
  isAuthenticated = false;

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.authService.isAuthenticated$.subscribe(isAuth => {
      this.isAuthenticated = isAuth;
    });
    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
    });
  }

  onLogout(): void {
    this.authService.logout();
    this.router.navigate(['/auth/login']);
  }

  onRefreshToken(): void {
    alert('Token refreshed! (UI mock)');
  }

  onRoleChange(): void {
    // For demo only.
    if (this.currentUser) {
      const user = { ...this.currentUser, role: this.currentUser.role };
      localStorage.setItem('user', JSON.stringify(user));
      window.location.reload();
    }
  }
}

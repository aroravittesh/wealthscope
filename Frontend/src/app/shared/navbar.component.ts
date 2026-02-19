// (Removed stray code above imports)
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
          <!-- Logo -->
          <div class="flex items-center gap-3 cursor-pointer group" routerLink="/dashboard">
            <div class="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-xl flex items-center justify-center transform group-hover:scale-110 group-hover:shadow-lg group-hover:shadow-blue-500/50 transition-all duration-300">
                <span class="text-lg font-bold text-white">F</span>
            </div>
            <span class="text-white font-bold text-xl hidden sm:block group-hover:text-blue-300 transition-colors">FinSight</span>
          </div>

          <!-- Navigation Links -->
          <div class="hidden md:flex items-center gap-8">
            <a routerLink="/dashboard" 
               routerLinkActive="text-blue-400" 
               [routerLinkActiveOptions]="{ exact: true }"
               class="text-slate-300 hover:text-white hover:scale-105 transition-all duration-300">Dashboard</a>
            <a routerLink="/portfolio" 
               routerLinkActive="text-blue-400"
               class="text-slate-300 hover:text-white hover:scale-105 transition-all duration-300">Portfolios</a>
            <a routerLink="/analytics" class="text-slate-300 hover:text-white hover:scale-105 transition-all duration-300">Analytics</a>
            <a href="#" class="text-slate-300 hover:text-white hover:scale-105 transition-all duration-300">Watchlist</a>
            <!-- ADMIN links -->
            <ng-container *ngIf="currentUser?.role === 'ADMIN'">
              <a routerLink="/admin" class="text-yellow-400 hover:text-white hover:scale-105 transition-all duration-300 font-bold">Admin Dashboard</a>
              <a routerLink="/admin/users" class="text-yellow-400 hover:text-white hover:scale-105 transition-all duration-300">User Management</a>
              <a routerLink="/admin/assets" class="text-yellow-400 hover:text-white hover:scale-105 transition-all duration-300">Asset Management</a>
            </ng-container>
          </div>

          <!-- Right Side - User Menu -->
          <div class="flex items-center gap-4">
            <!-- Notifications -->
            <button class="relative p-2 text-slate-300 hover:text-white transition hover:scale-110 duration-300">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path>
              </svg>
              <span class="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
            </button>

            <!-- User Profile Dropdown -->
            <div class="relative group">
              <button class="flex items-center gap-2 p-2 rounded-lg hover:bg-slate-700 transition group-hover:scale-105 duration-300">
                <div class="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center group-hover:shadow-lg group-hover:shadow-blue-500/50 transition-all duration-300">
                  <span class="text-xs font-bold text-white">{{ currentUser?.fullName?.charAt(0) || 'U' }}</span>
                </div>
                <span *ngIf="currentUser?.role" class="ml-2 px-2 py-0.5 rounded text-xs font-semibold"
                  [ngClass]="{
                    'bg-blue-900 text-blue-300 border border-blue-500': currentUser?.role === 'USER',
                    'bg-yellow-900 text-yellow-300 border border-yellow-500': currentUser?.role === 'ADMIN'
                  }">
                  {{ currentUser?.role }}
                </span>
                <svg class="w-4 h-4 text-slate-300 group-hover:text-blue-300 transition" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3"></path>
                </svg>
              </button>

              <!-- Dropdown Menu -->
              <div class="absolute right-0 mt-0 w-48 bg-slate-800 border border-slate-700 rounded-lg shadow-xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 transform group-hover:scale-100 scale-95 origin-top-right">
                <div class="p-4 border-b border-slate-700">
                  <p class="text-white font-semibold text-sm">{{ currentUser?.fullName }}</p>
                  <p class="text-slate-400 text-xs">{{ currentUser?.email }}</p>
                </div>
                    <a routerLink="/profile" class="block px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 transition text-sm">Profile Settings</a>
                <a href="#" class="block px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 transition text-sm">Preferences</a>
                <a href="#" class="block px-4 py-2 text-slate-300 hover:text-white hover:bg-slate-700 transition text-sm">Help & Support</a>
                                <a routerLink="/system-health" class="block px-4 py-2 text-green-300 hover:text-white hover:bg-slate-700 transition text-sm">System Health</a>
                <div class="border-t border-slate-700"></div>
                <button (click)="onLogout()" class="w-full text-left px-4 py-2 text-red-400 hover:text-red-300 hover:bg-slate-700 transition text-sm">Logout</button>
                <button type="button" class="w-full text-left px-4 py-2 text-blue-400 hover:text-blue-300 hover:bg-slate-700 transition text-sm" (click)="onRefreshToken()">Refresh Token (Mock)</button>
              </div>
            </div>
            <!-- Mock role switcher for demo/testing -->
            <div *ngIf="currentUser" class="ml-2">
              <select [(ngModel)]="currentUser.role" (change)="onRoleChange()" class="bg-slate-700 text-xs text-slate-300 rounded px-2 py-1 border border-slate-600">
                <option value="USER">USER</option>
                <option value="ADMIN">ADMIN</option>
              </select>
            </div>
          </div>

          <!-- Mobile Menu Button -->
          <button class="md:hidden p-2 text-slate-300 hover:text-white">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
            </svg>
          </button>
        </div>
      </div>
    </nav>
  `,
  styles: []
})
export class NavbarComponent implements OnInit {
  currentUser: User | null = null;

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.currentUser = this.authService.getCurrentUser();
  }

  onLogout(): void {
    this.authService.logout();
    this.router.navigate(['/auth/login']);
  }

  onRefreshToken(): void {
    // Mock UI feedback for token refresh
    alert('Token refreshed! (UI mock)');
  }

  onRoleChange(): void {
    // For demo: update role in localStorage and reload
    if (this.currentUser) {
      const user = { ...this.currentUser, role: this.currentUser.role };
      localStorage.setItem('currentUser', JSON.stringify(user));
      window.location.reload();
    }
  }
}

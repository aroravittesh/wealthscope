// import { Component, OnInit } from '@angular/core';
// import { CommonModule } from '@angular/common';
// import { RouterModule, Router } from '@angular/router';
// import { AuthService } from '../services/auth.service';
// import { User } from '../models';

// @Component({
//   selector: 'app-navbar',
//   standalone: true,
//   imports: [CommonModule, RouterModule],
//   template: `
//     <nav class="bg-slate-800 bg-opacity-50 backdrop-filter backdrop-blur-lg border-b border-slate-700 sticky top-0 z-50 transition-all duration-300 hover:bg-opacity-70">
//       <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
//         <div class="flex justify-between items-center h-16">
//           <div class="flex items-center gap-3 cursor-pointer group" [routerLink]="isAuthenticated ? '/dashboard' : '/'">
//             <img
//               src="/aurex.png"
//               alt="Aurex"
//               class="h-10 sm:h-9 w-auto max-w-[9rem] object-contain object-left transition-transform duration-300 group-hover:scale-105"
//             />
//           </div>

//           <div *ngIf="isAuthenticated; else publicNav" class="hidden md:flex items-center gap-8">
//             <a routerLink="/dashboard" routerLinkActive="text-blue-400" [routerLinkActiveOptions]="{ exact: true }" class="text-slate-300 hover:text-white transition-all duration-300">Dashboard</a>
//             <a routerLink="/portfolio" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Portfolios</a>
//             <a routerLink="/analytics" class="text-slate-300 hover:text-white transition-all duration-300">Analytics</a>
//             <a routerLink="/nebula" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Nebula</a>
//             <a *ngIf="currentUser?.role === 'ADMIN'" routerLink="/admin" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Admin</a>
//             <a class="text-slate-500 cursor-not-allowed">Watchlist</a>
//           </div>
//           <ng-template #publicNav></ng-template>

//           <div *ngIf="isAuthenticated; else guestActions" class="flex items-center gap-3">
//             <a routerLink="/profile" class="text-slate-300 hover:text-white text-sm">Profile</a>
//             <button type="button" (click)="onLogout()" class="bg-red-600 hover:bg-red-700 text-white text-sm font-semibold px-4 py-2 rounded-lg">
//               Logout
//             </button>
//           </div>
//           <ng-template #guestActions>
//             <div class="flex items-center gap-3">
//               <a routerLink="/auth/login" class="text-slate-300 hover:text-white font-semibold">Login</a>
//               <a routerLink="/auth/signup" class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-semibold px-4 py-2 rounded-lg transition">
//                 Sign Up
//               </a>
//             </div>
//           </ng-template>

//         </div>
//       </div>
//     </nav>
//   `,
//   styles: []
// })
// export class NavbarComponent implements OnInit {
//   currentUser: User | null = null;
//   isAuthenticated = false;

//   constructor(
//     private authService: AuthService,
//     private router: Router
//   ) {}

//   ngOnInit(): void {
//     this.authService.isAuthenticated$.subscribe(isAuth => {
//       this.isAuthenticated = isAuth;
//     });
//     this.authService.currentUser$.subscribe(user => {
//       this.currentUser = user;
//     });
//   }

//   onLogout(): void {
//     this.authService.logout();
//     this.router.navigate(['/auth/login']);
//   }
// }
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { User } from '../models';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <nav class="bg-slate-800 bg-opacity-50 backdrop-filter backdrop-blur-lg border-b border-slate-700 sticky top-0 z-50 transition-all duration-300 hover:bg-opacity-70">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">

          <div
            class="flex items-center gap-3 cursor-pointer group"
            [routerLink]="isAuthenticated ? '/dashboard' : '/'"
          >
            <!-- Animated Logo Container -->
            <div class="flex items-center gap-2 transition-transform duration-300 group-hover:scale-105">
              <!-- Geometric Icon with Moving Pattern -->
              <div class="relative flex items-center justify-center h-10 w-10 rounded-xl shadow-lg shadow-blue-500/30 overflow-hidden shrink-0">
                <div class="absolute inset-0 bg-gradient-to-r from-blue-400 via-indigo-500 to-purple-600 animate-gradient-x opacity-90"></div>
                <div class="absolute inset-[2px] bg-slate-800 rounded-lg z-0"></div>
                <svg class="relative z-10 w-6 h-6 text-transparent animate-spin-slow drop-shadow-[0_0_8px_rgba(59,130,246,0.8)]" style="fill: url(#logoGradient);" viewBox="0 0 24 24">
                  <defs>
                    <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" stop-color="#60A5FA" />
                      <stop offset="50%" stop-color="#818CF8" />
                      <stop offset="100%" stop-color="#C084FC" />
                    </linearGradient>
                  </defs>
                  <path d="M12 2L3 21h4.5l1.5-4h6l1.5 4H21L12 2zm0 4.5l2.5 7h-5L12 6.5z" stroke="url(#logoGradient)" stroke-width="1" stroke-linejoin="round" fill="url(#logoGradient)"/>
                </svg>
              </div>
              
              <!-- Text with Moving Gradient Pattern -->
              <span class="text-2xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-blue-400 via-purple-400 to-blue-400 animate-gradient-x drop-shadow-sm hidden sm:block">
                Aurex
              </span>
            </div>
          </div>

          <div *ngIf="isAuthenticated; else publicNav" class="hidden md:flex items-center gap-8">
            <a routerLink="/dashboard" routerLinkActive="text-blue-400" [routerLinkActiveOptions]="{ exact: true }" class="text-slate-300 hover:text-white transition-all duration-300">Dashboard</a>
            <a routerLink="/portfolio" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Portfolios</a>
            <a routerLink="/analytics" class="text-slate-300 hover:text-white transition-all duration-300">Analytics</a>
            <a routerLink="/nebula" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Nebula</a>
            <a *ngIf="currentUser?.role === 'ADMIN'" routerLink="/admin" routerLinkActive="text-blue-400" class="text-slate-300 hover:text-white transition-all duration-300">Admin</a>
            <a class="text-slate-500 cursor-not-allowed">Watchlist</a>
          </div>

          <ng-template #publicNav></ng-template>

          <div *ngIf="isAuthenticated; else guestActions" class="flex items-center gap-3">
            <a routerLink="/profile" class="text-slate-300 hover:text-white text-sm">Profile</a>
            <button
              type="button"
              (click)="onLogout()"
              class="bg-red-600 hover:bg-red-700 text-white text-sm font-semibold px-4 py-2 rounded-lg"
            >
              Logout
            </button>
          </div>

          <ng-template #guestActions>
            <div class="flex items-center gap-3">
              <a routerLink="/auth/login" class="text-slate-300 hover:text-white font-semibold">Login</a>
              <a
                routerLink="/auth/signup"
                class="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-semibold px-4 py-2 rounded-lg transition"
              >
                Sign Up
              </a>
            </div>
          </ng-template>

        </div>
      </div>
    </nav>
  `,
  styles: [`
    @keyframes gradient-x {
      0%, 100% {
        background-size: 200% 200%;
        background-position: left center;
      }
      50% {
        background-size: 200% 200%;
        background-position: right center;
      }
    }
    .animate-gradient-x {
      animation: gradient-x 3s ease infinite;
    }

    @keyframes spin-slow {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }
    .animate-spin-slow {
      animation: spin-slow 12s linear infinite;
    }


    .logo-blend {
      mix-blend-mode: screen;
      opacity: 0.95;
      filter: brightness(1.15) contrast(1.08);
    }

    .logo-blend:hover {
      opacity: 1;
      filter: brightness(1.3) contrast(1.12);
    }
  `]
})
export class NavbarComponent implements OnInit {
  currentUser: User | null = null;
  isAuthenticated = false;

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.authService.isAuthenticated$.subscribe((isAuth: boolean) => {
      this.isAuthenticated = isAuth;
    });

    this.authService.currentUser$.subscribe((user: User | null) => {
      this.currentUser = user;
    });
  }

  onLogout(): void {
    this.authService.logout();
    this.router.navigate(['/auth/login']);
  }
}
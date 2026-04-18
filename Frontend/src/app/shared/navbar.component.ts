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
            <img
  src="/aurex.png"
  alt="Aurex"
  class="h-16 sm:h-14 md:h-16 w-auto max-w-[14rem] object-contain object-left transition-transform duration-300 group-hover:scale-105"
/>
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
}
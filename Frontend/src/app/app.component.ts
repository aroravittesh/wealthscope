import { Component, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { CommonModule } from '@angular/common';
import { NavbarComponent } from './shared/navbar.component';
import { AuthService } from './services/auth.service';
import { filter } from 'rxjs/operators';
import { ChatbotLauncherComponent } from './features/chatbot/chatbot-launcher.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, CommonModule, NavbarComponent, ChatbotLauncherComponent],
  template: `
    <div
      class="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex flex-col"
    >
      <app-navbar *ngIf="showNavbar"></app-navbar>

      <main class="flex-1 min-h-0 w-full" [class]="showNavbar ? ' mx-auto  ' : ''">
        <router-outlet></router-outlet>
      </main>
      <app-chatbot-launcher></app-chatbot-launcher>
    </div>
  `,
  styles: []
})
export class AppComponent implements OnInit {
  showNavbar = false;
  private isAuthenticated = false;
  private currentUrl = '/';

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.authService.isAuthenticated$.subscribe(isAuth => {
      this.isAuthenticated = isAuth;
      this.updateNavbarVisibility();
    });

    this.currentUrl = this.router.url || '/';
    this.updateNavbarVisibility();

    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .subscribe(event => {
        this.currentUrl = (event as NavigationEnd).urlAfterRedirects || '/';
        this.updateNavbarVisibility();
      });
  }

  private updateNavbarVisibility(): void {
    const path = (this.currentUrl || '/').split('?')[0].split('#')[0];
    const isLandingPage = path === '/' || path === '';
    const isAuthPage = path === '/auth/login' || path === '/auth/signup';
    // Full-bleed auth screens: logo lives in the page, not the global nav.
    if (isAuthPage) {
      this.showNavbar = false;
      return;
    }
    this.showNavbar = this.isAuthenticated || isLandingPage;
  }
}

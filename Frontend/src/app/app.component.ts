import { Component, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { CommonModule } from '@angular/common';
import { NavbarComponent } from './shared/navbar.component';
import { AuthService } from './services/auth.service';
import { filter } from 'rxjs/operators';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, CommonModule, NavbarComponent],
  template: `
    <div class="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <app-navbar *ngIf="showNavbar"></app-navbar>
      
      <main [class]="showNavbar ? ' mx-auto  ' : ''">
        <router-outlet></router-outlet>
      </main>
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
    // Keep navbar visible on public landing page so login/signup is always reachable.
    const isLandingPage = this.currentUrl === '/';
    this.showNavbar = this.isAuthenticated || isLandingPage;
  }
}

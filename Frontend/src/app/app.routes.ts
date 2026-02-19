import { Routes } from '@angular/router';
import { AuthGuard } from './guards/auth.guard';
import { LoginComponent } from './features/auth/login/login.component';
import { RegisterComponent } from './features/auth/register/register.component';
import { DashboardOverviewComponent } from './features/dashboard/overview/overview.component';
import { LandingComponent } from './features/landing/landing.component';
import { PaymentComponent } from './features/payment/payment.component';
import { FeaturesComponent } from './features/pages/features.component';
import { PricingComponent } from './features/pages/pricing.component';
import { SecurityComponent } from './features/pages/security.component';
import { AboutComponent } from './features/pages/about.component';
import { BlogComponent } from './features/pages/blog.component';
import { CareersComponent } from './features/pages/careers.component';
import { DocsComponent } from './features/pages/docs.component';
import { ApiComponent } from './features/pages/api.component';
import { SupportComponent } from './features/pages/support.component';
import { PrivacyComponent } from './features/pages/privacy.component';
import { TermsComponent } from './features/pages/terms.component';
import { CookiesComponent } from './features/pages/cookies.component';
import { UserProfileComponent } from './features/user-profile.component';
import { AdminDashboardComponent } from './features/admin-dashboard.component';
import { ReportingAnalyticsComponent } from './features/reporting-analytics.component';
import { SystemHealthComponent } from './features/system-health.component';
import { DevopsTestingComponent } from './features/devops-testing.component';


export const routes: Routes = [
    {
      path: 'devops-testing',
      component: DevopsTestingComponent
    },
    {
      path: 'analytics',
      component: ReportingAnalyticsComponent
    },
    {
      path: 'system-health',
      component: SystemHealthComponent
    },
    {
      path: 'admin',
      component: AdminDashboardComponent
    },
  {
    path: '',
    component: LandingComponent
  },
  {
    path: 'auth',
    children: [
      { path: 'login', component: LoginComponent },
      { path: 'register', component: RegisterComponent }
    ]
  },
  {
    path: 'payment',
    component: PaymentComponent
  },
  {
    path: 'dashboard',
    canActivate: [AuthGuard],
    component: DashboardOverviewComponent
  },
  {
    path: 'portfolio',
    canActivate: [AuthGuard],
    children: []
  },
  // Product routes
  {
    path: 'features',
    component: FeaturesComponent
  },
  {
    path: 'pricing',
    component: PricingComponent
  },
  {
    path: 'security',
    component: SecurityComponent
  },
  // Company routes
  {
    path: 'about',
    component: AboutComponent
  },
  {
    path: 'blog',
    component: BlogComponent
  },
  {
    path: 'careers',
    component: CareersComponent
  },
  // Resources routes
  {
    path: 'docs',
    component: DocsComponent
  },
  {
    path: 'api',
    component: ApiComponent
  },
  {
    path: 'support',
    component: SupportComponent
  },
  // Legal routes
  {
    path: 'privacy',
    component: PrivacyComponent
  },
  {
    path: 'terms',
    component: TermsComponent
  },
  {
    path: 'cookies',
    component: CookiesComponent
  },
  {
    path: 'profile',
    component: UserProfileComponent
  }
];

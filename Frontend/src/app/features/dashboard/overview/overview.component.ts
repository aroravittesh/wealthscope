import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { PortfolioService } from '../../../services/portfolio.service';
import { AuthService } from '../../../services/auth.service';
import { Portfolio, DashboardMetrics } from '../../../models';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

@Component({
  selector: 'app-dashboard-overview',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './overview.component.html',
  styleUrl: './overview.component.scss'
})
export class DashboardOverviewComponent implements OnInit, OnDestroy {
  portfolios: Portfolio[] = [];
  metrics: DashboardMetrics | null = null;
  currentUser: any = null;
  animatedPortfolioValue: string = '0.00';
  animatedInvested: string = '0.00';
  animatedProfitLoss: string = '0.00';
  animatedProfitLossPercent: string = '0.00';
  liveChange: string = '+0.00';
  private destroy$ = new Subject<void>();
  private intervalId: any;

  constructor(
    private portfolioService: PortfolioService,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.currentUser = this.authService.getCurrentUser();
    this.loadPortfolios();
    this.simulateLiveMetrics();
  }

  simulateLiveMetrics(): void {
    // Animate numbers and simulate live value changes
    this.intervalId = setInterval(() => {
      // Fake values for demo
      const base = Math.random() * 10000 + 5000;
      const invested = base - Math.random() * 2000;
      const profitLoss = base - invested;
      const percent = (profitLoss / invested) * 100;
      const live = (Math.random() * 100 - 50);
      this.animatedPortfolioValue = base.toFixed(2);
      this.animatedInvested = invested.toFixed(2);
      this.animatedProfitLoss = profitLoss.toFixed(2);
      this.animatedProfitLossPercent = percent.toFixed(2);
      this.liveChange = (live > 0 ? '+' : '') + live.toFixed(2);
    }, 1200);
  }

  private loadPortfolios(): void {
    this.portfolioService.getPortfolios()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (portfolios) => {
          this.portfolios = portfolios;
        },
        error: (err) => console.error('Error loading portfolios', err)
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }
}

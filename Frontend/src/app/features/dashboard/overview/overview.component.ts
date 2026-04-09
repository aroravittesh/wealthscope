import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { PortfolioService } from '../../../services/portfolio.service';
import { AuthService } from '../../../services/auth.service';
import { Portfolio, DashboardMetrics } from '../../../models';
import { Subject, forkJoin, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
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
  private basePortfolios: Portfolio[] = [];
  metrics: DashboardMetrics | null = null;
  currentUser: any = null;
  loading = false;
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
    this.intervalId = setInterval(() => {
      this.applyLiveShift();
    }, 1200);
  }

  private loadPortfolios(): void {
    this.loading = true;
    this.portfolioService.getPortfolios()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (portfolios) => {
          if (!portfolios.length) {
            this.portfolios = [];
            this.metrics = this.buildMetrics([]);
            this.loading = false;
            return;
          }

          const holdingsRequests = portfolios.map(portfolio =>
            this.portfolioService.getHoldings(portfolio.id).pipe(
              catchError(() => of([])),
              map(holdings => {
                const invested = holdings.reduce((sum, h) => sum + (Number(h.quantity) * Number(h.avgPrice)), 0);
                return {
                  ...portfolio,
                  totalValue: invested,
                  totalInvested: invested,
                  totalProfitLoss: 0,
                  profitLossPercentage: 0
                } as Portfolio;
              })
            )
          );

          forkJoin(holdingsRequests).subscribe({
            next: enriched => {
              this.basePortfolios = enriched.map(p => ({ ...p }));
              this.applyLiveShift();
              this.loading = false;
            },
            error: () => {
              this.basePortfolios = portfolios.map(p => ({ ...p }));
              this.applyLiveShift();
              this.loading = false;
            }
          });
        },
        error: (err) => {
          console.error('Error loading portfolios', err);
          this.loading = false;
          this.portfolios = [];
          this.metrics = this.buildMetrics([]);
        }
      });
  }

  private buildMetrics(portfolios: Portfolio[]): DashboardMetrics {
    const totalPortfolioValue = portfolios.reduce((sum, p) => sum + Number(p.totalValue ?? 0), 0);
    const totalInvested = portfolios.reduce((sum, p) => sum + Number(p.totalInvested ?? p.totalValue ?? 0), 0);
    const totalProfitLoss = totalPortfolioValue - totalInvested;
    const profitLossPercentage = totalInvested > 0 ? (totalProfitLoss / totalInvested) * 100 : 0;

    return {
      totalPortfolioValue,
      totalInvested,
      totalProfitLoss,
      profitLossPercentage,
      assetsCount: 0,
      portfoliosCount: portfolios.length,
      topPerformers: [],
      allocationData: {}
    };
  }

  private applyLiveShift(): void {
    if (!this.basePortfolios.length) {
      this.portfolios = [];
      this.metrics = this.buildMetrics([]);
      this.animatedPortfolioValue = '0.00';
      this.animatedInvested = '0.00';
      this.animatedProfitLoss = '0.00';
      this.animatedProfitLossPercent = '0.00';
      this.liveChange = '+0.00%';
      return;
    }

    // Global market movement so dashboard cards and portfolio card P/L stay in sync.
    const shift = (Math.random() * 0.16) - 0.08; // -8% to +8%
    this.liveChange = `${shift >= 0 ? '+' : ''}${(shift * 100).toFixed(2)}%`;

    this.portfolios = this.basePortfolios.map(p => {
      const invested = Number(p.totalInvested ?? p.totalValue ?? 0);
      const value = invested * (1 + shift);
      const pnl = value - invested;
      return {
        ...p,
        totalInvested: invested,
        totalValue: value,
        totalProfitLoss: pnl,
        profitLossPercentage: shift * 100
      };
    });

    this.metrics = this.buildMetrics(this.portfolios);
    this.animatedPortfolioValue = (this.metrics.totalPortfolioValue ?? 0).toFixed(2);
    this.animatedInvested = (this.metrics.totalInvested ?? 0).toFixed(2);
    this.animatedProfitLoss = (this.metrics.totalProfitLoss ?? 0).toFixed(2);
    this.animatedProfitLossPercent = (this.metrics.profitLossPercentage ?? 0).toFixed(2);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }
}

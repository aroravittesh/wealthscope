import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { Subject, forkJoin, of } from 'rxjs';
import { catchError, map, takeUntil } from 'rxjs/operators';

import { PortfolioService } from '../../../services/portfolio.service';
import { Portfolio } from '../../../models';

@Component({
  selector: 'app-portfolio-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './portfolio-list.component.html'
})
export class PortfolioListComponent implements OnInit, OnDestroy {
  portfolios: Portfolio[] = [];
  private basePortfolios: Portfolio[] = [];
  loading = false;
  errorMessage: string | null = null;
  actionError: string | null = null;
  creating = false;
  newPortfolioName = '';
  renameTarget: Portfolio | null = null;
  renameValue = '';
  deletingId: string | null = null;
  liveChange = '+0.00%';
  private intervalId: any;

  private destroy$ = new Subject<void>();

  constructor(
    private portfolioService: PortfolioService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadPortfolios();
    this.startLiveSimulation();
  }

  private loadPortfolios(): void {
    this.loading = true;
    this.portfolioService
      .getPortfolios()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: portfolios => {
          if (!portfolios.length) {
            this.portfolios = [];
            return;
          }

          const requests = portfolios.map(portfolio =>
            this.portfolioService.getHoldings(portfolio.id).pipe(
              map(holdings => ({
                ...portfolio,
                totalValue: holdings.reduce((sum, h) => sum + (Number(h.quantity) * Number(h.avgPrice)), 0)
              })),
              catchError(() => of({ ...portfolio, totalValue: 0 }))
            )
          );

          forkJoin(requests).subscribe({
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
        error: err => {
          this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load portfolios';
          this.loading = false;
        },
        complete: () => {}
      });
  }

  createPortfolio(): void {
    const name = this.newPortfolioName.trim();
    if (!name) {
      this.actionError = 'Portfolio name is required';
      return;
    }
    this.creating = true;
    this.actionError = null;
    this.portfolioService.createPortfolio({ name }).pipe(takeUntil(this.destroy$)).subscribe({
      next: () => {
        this.creating = false;
        this.newPortfolioName = '';
      },
      error: err => {
        this.creating = false;
        this.actionError = err?.error?.message ?? err?.error ?? 'Failed to create portfolio';
      }
    });
  }

  startRename(portfolio: Portfolio): void {
    this.renameTarget = portfolio;
    this.renameValue = portfolio.name;
    this.actionError = null;
  }

  cancelRename(): void {
    this.renameTarget = null;
    this.renameValue = '';
  }

  saveRename(): void {
    if (!this.renameTarget) return;
    const targetId = this.renameTarget.id;
    const name = this.renameValue.trim();
    if (!name) {
      this.actionError = 'Portfolio name is required';
      return;
    }

    this.portfolioService.updatePortfolio(targetId, { name })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.portfolios = this.portfolios.map(p =>
            p.id === targetId ? { ...p, name } : p
          );
          this.cancelRename();
        },
        error: err => {
          this.actionError = err?.error?.message ?? err?.error ?? 'Failed to rename portfolio';
        }
      });
  }

  deletePortfolio(portfolio: Portfolio): void {
    if (!portfolio?.id) return;
    if (!confirm(`Delete portfolio "${portfolio.name}"?`)) return;
    this.deletingId = portfolio.id;
    this.actionError = null;
    this.portfolioService.deletePortfolio(portfolio.id).pipe(takeUntil(this.destroy$)).subscribe({
      next: () => {
        this.deletingId = null;
        if (this.renameTarget?.id === portfolio.id) {
          this.cancelRename();
        }
      },
      error: err => {
        this.deletingId = null;
        this.actionError = err?.error?.message ?? err?.error ?? 'Failed to delete portfolio';
      }
    });
  }

  openPortfolio(portfolio: Portfolio): void {
    if (!portfolio?.id) return;
    this.router.navigate(['/portfolio', portfolio.id, 'holdings']);
  }

  private startLiveSimulation(): void {
    this.intervalId = setInterval(() => {
      this.applyLiveShift();
    }, 1200);
  }

  private applyLiveShift(): void {
    if (!this.basePortfolios.length) {
      this.portfolios = [];
      this.liveChange = '+0.00%';
      return;
    }

    const shift = (Math.random() * 0.16) - 0.08; // -8% to +8%
    this.liveChange = `${shift >= 0 ? '+' : ''}${(shift * 100).toFixed(2)}%`;

    this.portfolios = this.basePortfolios.map(p => {
      const invested = Number(p.totalValue ?? 0);
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
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }
}


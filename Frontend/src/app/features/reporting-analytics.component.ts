import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { forkJoin, of, Subject } from 'rxjs';
import { catchError, map, takeUntil } from 'rxjs/operators';
import { PortfolioService } from '../services/portfolio.service';
import { Holding, Portfolio } from '../models';

@Component({
  selector: 'app-reporting-analytics',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <div class="max-w-7xl min-h-[calc(100vh-4rem)] mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-8">
        <div>
          <h1 class="text-3xl font-bold text-white">Analytics</h1>
          <p class="text-slate-400 mt-1">Live summary from your current portfolio holdings.</p>
        </div>
        <div class="flex gap-2">
          <button
            type="button"
            (click)="downloadCsv()"
            class="bg-emerald-600 hover:bg-emerald-700 text-white font-semibold px-4 py-2 rounded-lg transition"
            [disabled]="loading"
          >
            Download CSV
          </button>
          <button
            type="button"
            (click)="downloadPdf()"
            class="bg-purple-600 hover:bg-purple-700 text-white font-semibold px-4 py-2 rounded-lg transition"
            [disabled]="loading"
          >
            Download PDF
          </button>
          <a routerLink="/portfolio" class="bg-blue-600 hover:bg-blue-700 text-white font-semibold px-4 py-2 rounded-lg transition w-fit">
            Manage Portfolios
          </a>
        </div>
      </div>

      <div *ngIf="loading" class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div *ngFor="let i of [1,2,3]" class="bg-slate-800/70 rounded-xl p-6 border border-slate-700 animate-pulse">
          <div class="h-4 w-32 bg-slate-700 rounded mb-3"></div>
          <div class="h-8 w-24 bg-slate-700 rounded"></div>
        </div>
      </div>

      <div *ngIf="!loading && errorMessage" class="bg-red-900/30 border border-red-500/50 rounded-xl p-4 text-red-200">
        {{ errorMessage }}
      </div>

      <ng-container *ngIf="!loading && !errorMessage">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
            <p class="text-slate-400 text-sm">Total Portfolio Value</p>
            <p class="text-white text-3xl font-bold mt-2">\${{ totalValue.toFixed(2) }}</p>
          </div>
          <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
            <p class="text-slate-400 text-sm">Total Invested</p>
            <p class="text-white text-3xl font-bold mt-2">\${{ totalInvested.toFixed(2) }}</p>
          </div>
          <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
            <p class="text-slate-400 text-sm">Profit / Loss</p>
            <p class="text-3xl font-bold mt-2" [ngClass]="profitLoss >= 0 ? 'text-green-400' : 'text-red-400'">
              \${{ profitLoss.toFixed(2) }}
            </p>
            <p class="text-xs mt-1" [ngClass]="profitLossPercent >= 0 ? 'text-green-300' : 'text-red-300'">
              {{ profitLossPercent.toFixed(2) }}%
            </p>
          </div>
        </div>

        <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700 mb-6">
          <h2 class="text-xl font-semibold text-white mb-4">Asset Allocation</h2>
          <div *ngIf="allocationRows.length === 0" class="text-slate-400 text-sm">
            No holdings found. Add holdings to view allocation.
          </div>
          <div *ngFor="let row of allocationRows" class="mb-3">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-slate-200 font-medium">{{ row.symbol }}</span>
              <span class="text-slate-300">{{ row.percent.toFixed(2) }}%</span>
            </div>
            <div class="w-full h-2 bg-slate-700 rounded-full overflow-hidden">
              <div class="h-2 bg-gradient-to-r from-blue-500 to-purple-600" [style.width.%]="row.percent"></div>
            </div>
          </div>
        </div>

        <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
          <h2 class="text-xl font-semibold text-white mb-4">Portfolio Breakdown</h2>
          <div *ngIf="portfolioRows.length === 0" class="text-slate-400 text-sm">
            No portfolios found.
          </div>
          <div *ngFor="let row of portfolioRows" class="flex items-center justify-between py-3 border-b border-slate-700 last:border-0">
            <div>
              <p class="text-white font-semibold">{{ row.name }}</p>
              <p class="text-slate-400 text-xs">{{ row.holdingsCount }} holdings</p>
            </div>
            <div class="text-right">
              <p class="text-white font-semibold">\${{ row.value.toFixed(2) }}</p>
              <p class="text-xs text-slate-400">{{ row.share.toFixed(2) }}% of total</p>
            </div>
          </div>
        </div>
      </ng-container>
    </div>
  `,
  styles: []
})
export class ReportingAnalyticsComponent implements OnInit, OnDestroy {
  loading = false;
  errorMessage: string | null = null;

  totalValue = 0;
  totalInvested = 0;
  profitLoss = 0;
  profitLossPercent = 0;

  allocationRows: Array<{ symbol: string; percent: number }> = [];
  portfolioRows: Array<{ name: string; value: number; share: number; holdingsCount: number }> = [];

  private destroy$ = new Subject<void>();

  constructor(private portfolioService: PortfolioService) {}

  ngOnInit(): void {
    this.loadAnalytics();
  }

  private loadAnalytics(): void {
    this.loading = true;
    this.errorMessage = null;

    this.portfolioService.getPortfolios().pipe(takeUntil(this.destroy$)).subscribe({
      next: portfolios => {
        if (!portfolios.length) {
          this.resetAnalytics();
          this.loading = false;
          return;
        }

        const requests = portfolios.map(portfolio =>
          this.portfolioService.getHoldings(portfolio.id).pipe(
            map(holdings => ({ portfolio, holdings })),
            catchError(() => of({ portfolio, holdings: [] as Holding[] }))
          )
        );

        forkJoin(requests).pipe(takeUntil(this.destroy$)).subscribe({
          next: rows => {
            this.buildAnalytics(rows);
            this.loading = false;
          },
          error: () => {
            this.errorMessage = 'Failed to load analytics';
            this.loading = false;
          }
        });
      },
      error: err => {
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load analytics';
        this.loading = false;
      }
    });
  }

  private buildAnalytics(rows: Array<{ portfolio: Portfolio; holdings: Holding[] }>): void {
    const symbolTotals = new Map<string, number>();
    const portfolioStats: Array<{ name: string; value: number; holdingsCount: number }> = [];

    let total = 0;

    rows.forEach(({ portfolio, holdings }) => {
      const portfolioValue = holdings.reduce((sum, h) => sum + (Number(h.quantity) * Number(h.avgPrice)), 0);
      total += portfolioValue;
      portfolioStats.push({
        name: portfolio.name,
        value: portfolioValue,
        holdingsCount: holdings.length
      });

      holdings.forEach(h => {
        const value = Number(h.quantity) * Number(h.avgPrice);
        symbolTotals.set(h.symbol, (symbolTotals.get(h.symbol) ?? 0) + value);
      });
    });

    this.totalValue = total;
    this.totalInvested = total;
    this.profitLoss = this.totalValue - this.totalInvested;
    this.profitLossPercent = this.totalInvested > 0 ? (this.profitLoss / this.totalInvested) * 100 : 0;

    this.allocationRows = Array.from(symbolTotals.entries())
      .map(([symbol, value]) => ({
        symbol,
        percent: total > 0 ? (value / total) * 100 : 0
      }))
      .sort((a, b) => b.percent - a.percent);

    this.portfolioRows = portfolioStats
      .map(p => ({
        ...p,
        share: total > 0 ? (p.value / total) * 100 : 0
      }))
      .sort((a, b) => b.value - a.value);
  }

  private resetAnalytics(): void {
    this.totalValue = 0;
    this.totalInvested = 0;
    this.profitLoss = 0;
    this.profitLossPercent = 0;
    this.allocationRows = [];
    this.portfolioRows = [];
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  downloadCsv(): void {
    const rows: string[] = [];
    rows.push('Section,Metric,Value');
    rows.push(`Summary,Total Portfolio Value,${this.totalValue.toFixed(2)}`);
    rows.push(`Summary,Total Invested,${this.totalInvested.toFixed(2)}`);
    rows.push(`Summary,Profit/Loss,${this.profitLoss.toFixed(2)}`);
    rows.push(`Summary,Profit/Loss %,${this.profitLossPercent.toFixed(2)}%`);
    rows.push('');
    rows.push('Allocation,Symbol,Percent');
    this.allocationRows.forEach(row => {
      rows.push(`Allocation,${row.symbol},${row.percent.toFixed(2)}%`);
    });
    rows.push('');
    rows.push('Portfolio Breakdown,Name,Value,Share,Holdings Count');
    this.portfolioRows.forEach(row => {
      rows.push(`Portfolio,${row.name},${row.value.toFixed(2)},${row.share.toFixed(2)}%,${row.holdingsCount}`);
    });

    const blob = new Blob([rows.join('\n')], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `wealthscope-analytics-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  downloadPdf(): void {
    const lines: string[] = [];
    lines.push('WealthScope Analytics Report');
    lines.push(`Generated: ${new Date().toLocaleString()}`);
    lines.push('');
    lines.push(`Total Portfolio Value: $${this.totalValue.toFixed(2)}`);
    lines.push(`Total Invested: $${this.totalInvested.toFixed(2)}`);
    lines.push(`Profit/Loss: $${this.profitLoss.toFixed(2)} (${this.profitLossPercent.toFixed(2)}%)`);
    lines.push('');
    lines.push('Asset Allocation:');
    if (!this.allocationRows.length) {
      lines.push('- No holdings found');
    } else {
      this.allocationRows.forEach(r => lines.push(`- ${r.symbol}: ${r.percent.toFixed(2)}%`));
    }
    lines.push('');
    lines.push('Portfolio Breakdown:');
    if (!this.portfolioRows.length) {
      lines.push('- No portfolios found');
    } else {
      this.portfolioRows.forEach(r =>
        lines.push(`- ${r.name}: $${r.value.toFixed(2)} | ${r.share.toFixed(2)}% | ${r.holdingsCount} holdings`)
      );
    }

    const win = window.open('', '_blank');
    if (!win) {
      return;
    }
    win.document.write(`
      <html>
        <head><title>WealthScope Analytics Report</title></head>
        <body style="font-family: Arial, sans-serif; padding: 24px; white-space: pre-wrap;">
${lines.join('\n')}
        </body>
      </html>
    `);
    win.document.close();
    win.focus();
    win.print();
  }
}

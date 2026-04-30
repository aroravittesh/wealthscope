import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { PortfolioService } from '../services/portfolio.service';
import { Portfolio, PortfolioSnapshot, PortfolioSnapshotCompareResponse, PortfolioSummary } from '../models';

@Component({
  selector: 'app-reporting-analytics',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  template: `
    <div class="max-w-7xl min-h-[calc(100vh-4rem)] mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-8">
        <div>
          <h1 class="text-3xl font-bold text-white">Analytics</h1>
          <p class="text-slate-400 mt-1">Analyze one portfolio at a time and manage report snapshots.</p>
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

      <div class="bg-slate-800/70 rounded-xl p-4 border border-slate-700 mb-6 grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div>
          <label class="block text-xs text-slate-400 mb-1">Portfolio</label>
          <select
            [(ngModel)]="selectedPortfolioId"
            (change)="onPortfolioChange()"
            class="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-slate-200"
          >
            <option value="" disabled>Select portfolio</option>
            <option *ngFor="let p of portfolios" [value]="p.id">{{ p.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-xs text-slate-400 mb-1">Snapshot History</label>
          <select
            [(ngModel)]="selectedSnapshotId"
            (change)="onSnapshotChange()"
            class="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-slate-200"
            [disabled]="loading || snapshots.length === 0"
          >
            <option value="">Live data (no snapshot)</option>
            <option *ngFor="let s of snapshots" [value]="s.id">
              {{ s.createdAt | date:'medium' }}
            </option>
          </select>
        </div>

        <div class="flex items-end gap-2">
          <button
            type="button"
            (click)="saveSnapshot()"
            class="bg-yellow-600 hover:bg-yellow-500 text-white font-semibold px-4 py-2 rounded-lg transition"
            [disabled]="loading || !selectedPortfolioId || savingSnapshot"
          >
            {{ savingSnapshot ? 'Saving...' : 'Save Snapshot' }}
          </button>
          <button
            type="button"
            (click)="refreshLive()"
            class="bg-slate-700 hover:bg-slate-600 text-white font-semibold px-4 py-2 rounded-lg transition"
            [disabled]="loading || !selectedPortfolioId"
          >
            Refresh Live
          </button>
        </div>
      </div>

      <div class="bg-slate-800/70 rounded-xl p-4 border border-slate-700 mb-6 grid grid-cols-1 lg:grid-cols-4 gap-3">
        <div>
          <label class="block text-xs text-slate-400 mb-1">Then (from snapshot)</label>
          <select
            [(ngModel)]="compareFromSnapshotId"
            class="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-slate-200"
            [disabled]="compareLoading || snapshots.length < 2"
          >
            <option value="" disabled>Select older snapshot</option>
            <option *ngFor="let s of snapshots" [value]="s.id">{{ s.createdAt | date:'medium' }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-slate-400 mb-1">Now (to snapshot)</label>
          <select
            [(ngModel)]="compareToSnapshotId"
            class="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-slate-200"
            [disabled]="compareLoading || snapshots.length < 2"
          >
            <option value="" disabled>Select newer snapshot</option>
            <option *ngFor="let s of snapshots" [value]="s.id">{{ s.createdAt | date:'medium' }}</option>
          </select>
        </div>
        <div class="flex items-end">
          <button
            type="button"
            (click)="runSnapshotCompare()"
            class="bg-cyan-600 hover:bg-cyan-500 text-white font-semibold px-4 py-2 rounded-lg transition w-full"
            [disabled]="compareLoading || !canRunCompare"
          >
            {{ compareLoading ? 'Comparing...' : 'Compare Snapshots' }}
          </button>
        </div>
        <div class="flex items-end text-xs text-slate-400">
          <span *ngIf="snapshots.length < 2">Save at least 2 snapshots to compare progression.</span>
          <span *ngIf="snapshots.length >= 2 && snapshotCompare">Then: {{ snapshotCompare.fromAt | date:'short' }} → Now: {{ snapshotCompare.toAt | date:'short' }}</span>
        </div>
      </div>

      <div *ngIf="snapshotCompare" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 mb-8">
        <div class="bg-slate-800/70 rounded-xl p-5 border border-slate-700">
          <p class="text-slate-400 text-xs uppercase tracking-wide">Value Change</p>
          <p class="mt-2 text-xl font-bold" [ngClass]="deltaClass(snapshotCompare.totalValueDelta.absolute)">
            {{ signedCurrency(snapshotCompare.totalValueDelta.absolute) }}
          </p>
          <p class="text-xs mt-1" [ngClass]="deltaClass(snapshotCompare.totalValueDelta.percent)">
            {{ signedPercent(snapshotCompare.totalValueDelta.percent) }}
          </p>
        </div>
        <div class="bg-slate-800/70 rounded-xl p-5 border border-slate-700">
          <p class="text-slate-400 text-xs uppercase tracking-wide">P/L Change</p>
          <p class="mt-2 text-xl font-bold" [ngClass]="deltaClass(snapshotCompare.profitLossDelta.absolute)">
            {{ signedCurrency(snapshotCompare.profitLossDelta.absolute) }}
          </p>
          <p class="text-xs mt-1" [ngClass]="deltaClass(snapshotCompare.profitLossDelta.percent)">
            {{ signedPercent(snapshotCompare.profitLossDelta.percent) }}
          </p>
        </div>
        <div class="bg-slate-800/70 rounded-xl p-5 border border-slate-700">
          <p class="text-slate-400 text-xs uppercase tracking-wide">Diversification</p>
          <p class="mt-2 text-xl font-bold" [ngClass]="deltaClass(snapshotCompare.diversificationDelta.absolute)">
            {{ signedNumber(snapshotCompare.diversificationDelta.absolute) }}
          </p>
          <p class="text-xs mt-1" [ngClass]="deltaClass(snapshotCompare.diversificationDelta.percent)">
            {{ signedPercent(snapshotCompare.diversificationDelta.percent) }}
          </p>
        </div>
        <div class="bg-slate-800/70 rounded-xl p-5 border border-slate-700">
          <p class="text-slate-400 text-xs uppercase tracking-wide">Volatility</p>
          <p class="mt-2 text-xl font-bold" [ngClass]="deltaClass(snapshotCompare.volatilityDelta.absolute)">
            {{ signedNumber(snapshotCompare.volatilityDelta.absolute) }}
          </p>
          <p class="text-xs mt-1" [ngClass]="deltaClass(snapshotCompare.volatilityDelta.percent)">
            {{ signedPercent(snapshotCompare.volatilityDelta.percent) }}
          </p>
        </div>
      </div>

      <div *ngIf="loading" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div *ngFor="let i of [1,2,3]" class="bg-slate-800/70 rounded-xl p-6 border border-slate-700 animate-pulse">
            <div class="h-4 w-32 bg-slate-700 rounded mb-3"></div>
            <div class="h-8 w-24 bg-slate-700 rounded"></div>
          </div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div *ngFor="let j of [1,2]" class="bg-slate-800/70 rounded-xl p-6 border border-slate-700 animate-pulse">
            <div class="h-4 w-40 bg-slate-700 rounded mb-3"></div>
            <div class="h-3 w-full bg-slate-700 rounded"></div>
          </div>
        </div>
      </div>

      <div *ngIf="!loading && errorMessage" class="bg-red-900/30 border border-red-500/50 rounded-xl p-4 text-red-200">
        {{ errorMessage }}
      </div>

      <div *ngIf="!loading && !errorMessage && warningMessage" class="bg-amber-900/20 border border-amber-500/40 rounded-xl p-4 text-amber-100 mb-6">
        {{ warningMessage }}
      </div>

      <div *ngIf="!loading && !errorMessage" class="bg-slate-800/50 border border-slate-700 rounded-xl p-3 text-slate-300 text-sm mb-6">
        {{ reportContextLabel }}
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

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
          <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
            <p class="text-slate-400 text-sm">Diversification score</p>
            <p class="text-cyan-300 text-3xl font-bold mt-2">{{ diversificationScoreAgg.toFixed(1) }}</p>
            <p class="text-slate-500 text-xs mt-1">0–100 · higher = more balanced across holdings (value-weighted across portfolios)</p>
            <div class="w-full h-2 bg-slate-700 rounded-full overflow-hidden mt-3">
              <div
                class="h-2 bg-gradient-to-r from-cyan-600 to-teal-400 transition-all"
                [style.width.%]="diversificationScoreAgg"
              ></div>
            </div>
          </div>
          <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700">
            <p class="text-slate-400 text-sm">Volatility score</p>
            <p class="text-amber-300 text-3xl font-bold mt-2">{{ volatilityScoreAgg.toFixed(1) }}</p>
            <p class="text-slate-500 text-xs mt-1">0–100 · higher = more volatile (value-weighted across portfolios)</p>
            <div class="w-full h-2 bg-slate-700 rounded-full overflow-hidden mt-3">
              <div
                class="h-2 bg-gradient-to-r from-amber-700 to-orange-400 transition-all"
                [style.width.%]="volatilityScoreAgg"
              ></div>
            </div>
          </div>
        </div>

        <div class="bg-slate-800/70 rounded-xl p-6 border border-slate-700 mb-6">
          <h2 class="text-xl font-semibold text-white mb-4">Asset Allocation</h2>
          <div *ngIf="allocationRows.length === 0" class="text-slate-400 text-sm">
            No holdings found. Add holdings to view allocation.
          </div>
          <div *ngFor="let row of allocationRows" class="mb-3">
            <div class="flex justify-between text-sm mb-1 gap-2">
              <span class="text-slate-200 font-medium">{{ row.symbol }}</span>
              <span class="text-slate-300 shrink-0 text-right text-xs sm:text-sm">
                \${{ row.costBasis.toFixed(0) }} → \${{ row.value.toFixed(2) }} · {{ row.percent.toFixed(2) }}%
              </span>
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
          <div *ngFor="let row of portfolioRows" class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 py-3 border-b border-slate-700 last:border-0">
            <div>
              <p class="text-white font-semibold">{{ row.name }}</p>
              <p class="text-slate-400 text-xs">
                {{ row.holdingsCount }} holdings · Diversification {{ row.diversificationScore.toFixed(1) }} · Volatility
                {{ row.volatilityScore.toFixed(1) }}
              </p>
            </div>
            <div class="text-left sm:text-right">
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
  portfolios: Portfolio[] = [];
  selectedPortfolioId = '';
  selectedSnapshotId = '';
  compareFromSnapshotId = '';
  compareToSnapshotId = '';
  snapshots: PortfolioSnapshot[] = [];
  snapshotCompare: PortfolioSnapshotCompareResponse | null = null;
  compareLoading = false;
  savingSnapshot = false;

  loading = false;
  errorMessage: string | null = null;
  warningMessage: string | null = null;
  reportContextLabel = 'No portfolio selected.';

  totalValue = 0;
  totalInvested = 0;
  profitLoss = 0;
  profitLossPercent = 0;

  /** Value-weighted average across loaded portfolios (API scores are per-portfolio). */
  diversificationScoreAgg = 0;
  volatilityScoreAgg = 0;

  allocationRows: Array<{ symbol: string; value: number; costBasis: number; percent: number }> = [];
  portfolioRows: Array<{
    name: string;
    value: number;
    share: number;
    holdingsCount: number;
    diversificationScore: number;
    volatilityScore: number;
  }> = [];

  private destroy$ = new Subject<void>();

  constructor(private portfolioService: PortfolioService) {}

  ngOnInit(): void {
    this.loadPortfoliosAndInitialAnalytics();
  }

  private loadPortfoliosAndInitialAnalytics(): void {
    this.loading = true;
    this.errorMessage = null;
    this.warningMessage = null;
    this.reportContextLabel = 'Loading portfolios...';

    this.portfolioService.getPortfolios().pipe(takeUntil(this.destroy$)).subscribe({
      next: portfolios => {
        this.portfolios = portfolios;
        if (!portfolios.length) {
          this.resetAnalytics();
          this.loading = false;
          this.reportContextLabel = 'No portfolios available.';
          return;
        }
        this.selectedPortfolioId = portfolios[0].id;
        this.loadSnapshots();
        this.loadLiveAnalyticsForSelectedPortfolio();
      },
      error: err => {
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load analytics';
        this.loading = false;
      }
    });
  }

  onPortfolioChange(): void {
    this.selectedSnapshotId = '';
    this.compareFromSnapshotId = '';
    this.compareToSnapshotId = '';
    this.snapshotCompare = null;
    this.loadSnapshots();
    this.loadLiveAnalyticsForSelectedPortfolio();
  }

  onSnapshotChange(): void {
    if (!this.selectedSnapshotId) {
      this.loadLiveAnalyticsForSelectedPortfolio();
      return;
    }
    this.loadSnapshotById(this.selectedSnapshotId);
  }

  saveSnapshot(): void {
    if (!this.selectedPortfolioId) return;
    this.savingSnapshot = true;
    this.errorMessage = null;
    this.warningMessage = null;

    this.portfolioService.createPortfolioSnapshot(this.selectedPortfolioId).pipe(takeUntil(this.destroy$)).subscribe({
      next: snapshot => {
        this.savingSnapshot = false;
        this.selectedSnapshotId = snapshot.id;
        this.loadSnapshots();
        this.applySummary(snapshot.summary, this.getSelectedPortfolioName(), 0);
        this.reportContextLabel = `Snapshot view: ${this.getSelectedPortfolioName()} · ${snapshot.createdAt.toLocaleString()}`;
      },
      error: err => {
        this.savingSnapshot = false;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to save snapshot.';
      }
    });
  }

  refreshLive(): void {
    this.selectedSnapshotId = '';
    this.loadLiveAnalyticsForSelectedPortfolio();
  }

  private loadSnapshots(): void {
    if (!this.selectedPortfolioId) {
      this.snapshots = [];
      return;
    }
    this.portfolioService.getPortfolioSnapshots(this.selectedPortfolioId).pipe(takeUntil(this.destroy$)).subscribe({
      next: snapshots => {
        this.snapshots = [...snapshots].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
        this.setDefaultCompareSnapshotIds();
      },
      error: () => {
        this.snapshots = [];
        this.compareFromSnapshotId = '';
        this.compareToSnapshotId = '';
      }
    });
  }

  get canRunCompare(): boolean {
    return !!this.selectedPortfolioId &&
      !!this.compareFromSnapshotId &&
      !!this.compareToSnapshotId &&
      this.compareFromSnapshotId !== this.compareToSnapshotId;
  }

  runSnapshotCompare(): void {
    if (!this.canRunCompare || !this.selectedPortfolioId) {
      return;
    }
    this.compareLoading = true;
    this.errorMessage = null;

    this.portfolioService
      .getPortfolioSnapshotCompare(this.selectedPortfolioId, this.compareFromSnapshotId, this.compareToSnapshotId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: compare => {
          this.compareLoading = false;
          this.snapshotCompare = compare;
        },
        error: err => {
          this.compareLoading = false;
          this.snapshotCompare = null;
          this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to compare snapshots.';
        }
      });
  }

  private loadSnapshotById(snapshotId: string): void {
    if (!this.selectedPortfolioId || !snapshotId) return;
    this.errorMessage = null;
    this.warningMessage = null;

    const local = this.snapshots.find(s => s.id === snapshotId);
    if (local) {
      this.applySummary(local.summary, this.getSelectedPortfolioName(), 0);
      this.reportContextLabel = `Snapshot view: ${this.getSelectedPortfolioName()} · ${local.createdAt.toLocaleString()}`;
      return;
    }

    // Fallback: refresh snapshot list and resolve the selected entry from it.
    this.loading = true;
    this.portfolioService
      .getPortfolioSnapshots(this.selectedPortfolioId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: snapshots => {
          this.loading = false;
          this.snapshots = [...snapshots].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
          const resolved = this.snapshots.find(s => s.id === snapshotId);
          if (!resolved) {
            this.errorMessage = 'Snapshot not found.';
            return;
          }
          this.applySummary(resolved.summary, this.getSelectedPortfolioName(), 0);
          this.reportContextLabel = `Snapshot view: ${this.getSelectedPortfolioName()} · ${resolved.createdAt.toLocaleString()}`;
        },
        error: err => {
          this.loading = false;
          this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load snapshot.';
        }
      });
  }

  private loadLiveAnalyticsForSelectedPortfolio(): void {
    if (!this.selectedPortfolioId) {
      this.resetAnalytics();
      this.reportContextLabel = 'No portfolio selected.';
      return;
    }
    this.loading = true;
    this.errorMessage = null;
    this.warningMessage = null;

    this.portfolioService.getPortfolioSummary(this.selectedPortfolioId).pipe(takeUntil(this.destroy$)).subscribe({
      next: summary => {
        this.portfolioService.getHoldings(this.selectedPortfolioId).pipe(takeUntil(this.destroy$)).subscribe({
          next: holdings => {
            this.loading = false;
            this.applySummary(summary, this.getSelectedPortfolioName(), holdings.length);
            this.reportContextLabel = `Live view: ${this.getSelectedPortfolioName()}`;
          },
          error: () => {
            this.loading = false;
            this.applySummary(summary, this.getSelectedPortfolioName(), 0);
            this.warningMessage = 'Holdings count unavailable. Core analytics loaded.';
            this.reportContextLabel = `Live view: ${this.getSelectedPortfolioName()}`;
          }
        });
      },
      error: err => {
        this.loading = false;
        this.errorMessage = err?.error?.message ?? err?.error ?? 'Failed to load portfolio analytics.';
      }
    });
  }

  private applySummary(summary: PortfolioSummary, portfolioName: string, holdingsCount: number): void {
    const symbolTotals = new Map<string, { value: number; cost: number }>();
    for (const row of summary.assetAllocation) {
      const cur = symbolTotals.get(row.symbol) ?? { value: 0, cost: 0 };
      cur.value += row.value;
      cur.cost += row.costBasis;
      symbolTotals.set(row.symbol, cur);
    }

    this.totalValue = summary.totalPortfolioValue;
    this.totalInvested = summary.totalInvested;
    this.profitLoss = this.totalValue - this.totalInvested;
    this.profitLossPercent =
      this.totalInvested > 0 ? (this.profitLoss / this.totalInvested) * 100 : 0;

    this.diversificationScoreAgg = summary.diversificationScore;
    this.volatilityScoreAgg = summary.volatilityScore;

    this.allocationRows = Array.from(symbolTotals.entries())
      .map(([symbol, { value, cost }]) => ({
        symbol,
        value,
        costBasis: cost,
        percent: this.totalValue > 0 ? (value / this.totalValue) * 100 : 0
      }))
      .sort((a, b) => b.percent - a.percent);

    this.portfolioRows = [{
      name: portfolioName,
      value: summary.totalPortfolioValue,
      share: 100,
      holdingsCount,
      diversificationScore: summary.diversificationScore,
      volatilityScore: summary.volatilityScore
    }];
  }

  private resetAnalytics(): void {
    this.totalValue = 0;
    this.totalInvested = 0;
    this.profitLoss = 0;
    this.profitLossPercent = 0;
    this.diversificationScoreAgg = 0;
    this.volatilityScoreAgg = 0;
    this.allocationRows = [];
    this.portfolioRows = [];
    this.warningMessage = null;
  }

  private getSelectedPortfolioName(): string {
    const portfolio = this.portfolios.find(p => p.id === this.selectedPortfolioId);
    return portfolio?.name ?? 'Selected Portfolio';
  }

  private setDefaultCompareSnapshotIds(): void {
    if (this.snapshots.length < 2) {
      this.compareFromSnapshotId = '';
      this.compareToSnapshotId = '';
      this.snapshotCompare = null;
      return;
    }

    if (!this.compareFromSnapshotId) {
      this.compareFromSnapshotId = this.snapshots[this.snapshots.length - 1].id;
    }
    if (!this.compareToSnapshotId) {
      this.compareToSnapshotId = this.snapshots[0].id;
    }
  }

  deltaClass(value: number): string {
    if (value > 0) return 'text-green-400';
    if (value < 0) return 'text-red-400';
    return 'text-slate-200';
  }

  signedCurrency(value: number): string {
    return `${value >= 0 ? '+' : '-'}$${Math.abs(value).toFixed(2)}`;
  }

  signedPercent(value: number): string {
    return `${value >= 0 ? '+' : '-'}${Math.abs(value).toFixed(2)}%`;
  }

  signedNumber(value: number): string {
    return `${value >= 0 ? '+' : '-'}${Math.abs(value).toFixed(2)}`;
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  downloadCsv(): void {
    const rows: string[] = [];
    rows.push('Section,Metric,Value');
    rows.push(`Context,Data Source,${this.reportContextLabel}`);
    rows.push(`Summary,Total Portfolio Value,${this.totalValue.toFixed(2)}`);
    rows.push(`Summary,Total Invested,${this.totalInvested.toFixed(2)}`);
    rows.push(`Summary,Profit/Loss,${this.profitLoss.toFixed(2)}`);
    rows.push(`Summary,Profit/Loss %,${this.profitLossPercent.toFixed(2)}%`);
    rows.push(`Summary,Diversification score (aggregate),${this.diversificationScoreAgg.toFixed(2)}`);
    rows.push(`Summary,Volatility score (aggregate),${this.volatilityScoreAgg.toFixed(2)}`);
    rows.push('');
    rows.push('Allocation,Symbol,Cost Basis,Market Value,Percent');
    this.allocationRows.forEach(row => {
      rows.push(
        `Allocation,${row.symbol},${row.costBasis.toFixed(2)},${row.value.toFixed(2)},${row.percent.toFixed(2)}%`
      );
    });
    rows.push('');
    rows.push('Portfolio Breakdown,Name,Value,Share,Holdings Count,Diversification,Volatility');
    this.portfolioRows.forEach(row => {
      rows.push(
        `Portfolio,${row.name},${row.value.toFixed(2)},${row.share.toFixed(2)}%,${row.holdingsCount},${row.diversificationScore.toFixed(2)},${row.volatilityScore.toFixed(2)}`
      );
    });

    const blob = new Blob([rows.join('\n')], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `aurex-analytics-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  downloadPdf(): void {
    const lines: string[] = [];
    lines.push('Aurex Analytics Report');
    lines.push(`Generated: ${new Date().toLocaleString()}`);
    lines.push(`Context: ${this.reportContextLabel}`);
    lines.push('');
    lines.push(`Total Portfolio Value: $${this.totalValue.toFixed(2)}`);
    lines.push(`Total Invested: $${this.totalInvested.toFixed(2)}`);
    lines.push(`Profit/Loss: $${this.profitLoss.toFixed(2)} (${this.profitLossPercent.toFixed(2)}%)`);
    lines.push(
      `Diversification score (aggregate): ${this.diversificationScoreAgg.toFixed(1)} / 100`
    );
    lines.push(`Volatility score (aggregate): ${this.volatilityScoreAgg.toFixed(1)} / 100`);
    lines.push('');
    lines.push('Asset Allocation:');
    if (!this.allocationRows.length) {
      lines.push('- No holdings found');
    } else {
      this.allocationRows.forEach(r =>
        lines.push(
          `- ${r.symbol}: cost $${r.costBasis.toFixed(2)} → market $${r.value.toFixed(2)} (${r.percent.toFixed(2)}%)`
        )
      );
    }
    lines.push('');
    lines.push('Portfolio Breakdown:');
    if (!this.portfolioRows.length) {
      lines.push('- No portfolios found');
    } else {
      this.portfolioRows.forEach(r =>
        lines.push(
          `- ${r.name}: $${r.value.toFixed(2)} | ${r.share.toFixed(2)}% | ${r.holdingsCount} holdings | div ${r.diversificationScore.toFixed(1)} | vol ${r.volatilityScore.toFixed(1)}`
        )
      );
    }

    const win = window.open('', '_blank');
    if (!win) {
      return;
    }
    win.document.write(`
      <html>
        <head><title>Aurex Analytics Report</title></head>
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

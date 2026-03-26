import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { Holding, Portfolio } from '../../../models';
import { PortfolioService } from '../../../services/portfolio.service';

@Component({
  selector: 'app-holdings',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './holdings.component.html',
})
export class HoldingsComponent implements OnInit, OnDestroy {
  portfolioId: string | null = null;
  portfolio: Portfolio | null = null;

  holdings: Holding[] = [];
  loading = false;
  errorMessage: string | null = null;

  // Modal/form state
  showHoldingModal = false;
  isEditing = false;
  mutating = false;
  formErrorMessage: string | null = null;

  holdingForm: {
    symbol: string;
    assetType: string;
    quantity: number | null;
    avgPrice: number | null;
  } = {
    symbol: '',
    assetType: 'STOCK',
    quantity: null,
    avgPrice: null
  };

  private destroy$ = new Subject<void>();

  assetTypeOptions = [
    { value: 'STOCK', label: 'Stock' },
    { value: 'CRYPTO', label: 'Crypto' },
    { value: 'ETF', label: 'ETF' },
  ];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private portfolioService: PortfolioService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.pipe(takeUntil(this.destroy$)).subscribe(params => {
      const id = params.get('portfolioId');
      if (!id) {
        this.errorMessage = 'Missing portfolio id';
        return;
      }

      this.portfolioId = id;
      this.errorMessage = null;

      this.loadPortfolio(id);
      this.loadHoldings(id);
    });
  }

  private loadPortfolio(id: string): void {
    this.portfolioService
      .getPortfolioById(id)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: p => (this.portfolio = p),
        error: _err => {
          // Keep the holdings table functional even if portfolio header fails.
          this.portfolio = null;
        }
      });
  }

  private loadHoldings(id: string): void {
    this.loading = true;
    this.errorMessage = null;

    this.portfolioService
      .getHoldings(id)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: holdings => (this.holdings = holdings),
        error: err => {
          const backendMessage =
            err?.error?.message ?? err?.error ?? err?.message ?? 'Failed to load holdings';
          this.errorMessage = typeof backendMessage === 'string' ? backendMessage : 'Failed to load holdings';
          this.holdings = [];
          this.loading = false;
        },
        complete: () => {
          this.loading = false;
        }
      });
  }

  backToPortfolios(): void {
    this.router.navigate(['/portfolio']);
  }

  investedValue(holding: Holding): number {
    const quantity = Number(holding?.quantity ?? 0);
    const avgPrice = Number(holding?.avgPrice ?? 0);
    return quantity * avgPrice;
  }

  openAddModal(): void {
    this.isEditing = false;
    this.mutating = false;
    this.formErrorMessage = null;
    this.holdingForm = { symbol: '', assetType: 'STOCK', quantity: null, avgPrice: null };
    this.showHoldingModal = true;
  }

  openEditModal(holding: Holding): void {
    this.isEditing = true;
    this.mutating = false;
    this.formErrorMessage = null;
    this.holdingForm = {
      symbol: holding.symbol,
      assetType: (holding.assetType || '').toUpperCase(),
      quantity: holding.quantity,
      avgPrice: holding.avgPrice
    };
    this.showHoldingModal = true;
  }

  closeModal(): void {
    if (this.mutating) return;
    this.showHoldingModal = false;
    this.formErrorMessage = null;
  }

  private normalizeAndValidate(): {
    symbol: string;
    assetType: string;
    quantity: number;
    avgPrice: number;
  } | null {
    const symbol = (this.holdingForm.symbol || '').trim().toUpperCase();
    const assetType = (this.holdingForm.assetType || '').trim().toUpperCase();
    const quantity = Number(this.holdingForm.quantity ?? NaN);
    const avgPrice = Number(this.holdingForm.avgPrice ?? NaN);

    if (!symbol) {
      this.formErrorMessage = 'Symbol is required';
      return null;
    }
    if (!assetType) {
      this.formErrorMessage = 'Asset type is required';
      return null;
    }
    if (!['STOCK', 'CRYPTO', 'ETF'].includes(assetType)) {
      this.formErrorMessage = 'Asset type must be STOCK, CRYPTO, or ETF';
      return null;
    }
    if (!Number.isFinite(quantity) || quantity < 0) {
      this.formErrorMessage = 'Quantity must be a number >= 0';
      return null;
    }
    if (!Number.isFinite(avgPrice) || avgPrice < 0) {
      this.formErrorMessage = 'Avg price must be a number >= 0';
      return null;
    }

    return { symbol, assetType, quantity, avgPrice };
  }

  saveHolding(): void {
    if (!this.portfolioId) return;

    const normalized = this.normalizeAndValidate();
    if (!normalized) return;

    this.mutating = true;
    this.formErrorMessage = null;

    this.portfolioService
      .addOrUpdateHolding({
        portfolio_id: this.portfolioId,
        symbol: normalized.symbol,
        asset_type: normalized.assetType,
        quantity: normalized.quantity,
        avg_price: normalized.avgPrice
      })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.mutating = false;
          this.showHoldingModal = false;
          this.loadHoldings(this.portfolioId as string);
        },
        error: (err: any) => {
          this.mutating = false;
          const backendMessage =
            err?.error?.message ?? err?.error ?? err?.message ?? 'Failed to save holding';
          this.formErrorMessage = typeof backendMessage === 'string' ? backendMessage : 'Failed to save holding';
        }
      });
  }

  deleteHolding(holding: Holding): void {
    if (!holding?.id) return;

    const confirmed = confirm(`Delete ${holding.symbol} holding?`);
    if (!confirmed) return;

    if (!this.portfolioId) return;

    this.mutating = true;
    this.formErrorMessage = null;

    this.portfolioService
      .deleteHolding(holding.id)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.mutating = false;
          this.loadHoldings(this.portfolioId as string);
        },
        error: (err: any) => {
          this.mutating = false;
          const backendMessage =
            err?.error?.message ?? err?.error ?? err?.message ?? 'Failed to delete holding';
          this.formErrorMessage = typeof backendMessage === 'string' ? backendMessage : 'Failed to delete holding';
        }
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}


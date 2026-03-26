import { Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { PortfolioService } from '../../../services/portfolio.service';
import { Portfolio } from '../../../models';

@Component({
  selector: 'app-portfolio-list',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './portfolio-list.component.html'
})
export class PortfolioListComponent implements OnInit, OnDestroy {
  portfolios: Portfolio[] = [];
  loading = false;
  errorMessage: string | null = null;

  private destroy$ = new Subject<void>();

  constructor(private portfolioService: PortfolioService) {}

  ngOnInit(): void {
    this.loading = true;
    this.portfolioService
      .getPortfolios()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: portfolios => (this.portfolios = portfolios),
        error: err => {
          this.errorMessage = err?.error?.message || 'Failed to load portfolios';
          this.loading = false;
        },
        complete: () => {
          this.loading = false;
        }
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}


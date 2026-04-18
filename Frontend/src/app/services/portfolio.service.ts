import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { map, tap } from 'rxjs/operators';

import {
  AssetAllocationRow,
  DashboardMetrics,
  Holding,
  Portfolio,
  PortfolioSnapshot,
  PortfolioSummary
} from '../models';

@Injectable({
  providedIn: 'root'
})
export class PortfolioService {

  private portfolioApiUrl = `${environment.apiUrl}/portfolios`;
  private holdingsApiUrl = `${environment.apiUrl}/holdings`;

  private portfoliosSubject = new BehaviorSubject<Portfolio[]>([]);
  public portfolios$ = this.portfoliosSubject.asObservable();

  private currentPortfolioSubject = new BehaviorSubject<Portfolio | null>(null);
  public currentPortfolio$ = this.currentPortfolioSubject.asObservable();

  private metricsSubject = new BehaviorSubject<DashboardMetrics | null>(null);
  public metrics$ = this.metricsSubject.asObservable();

  constructor(private http: HttpClient) {}

  getPortfolios(): Observable<Portfolio[]> {
    return this.http.get<any[]>(this.portfolioApiUrl).pipe(
      map(rows => (rows || []).map(this.mapPortfolioFromApi)),
      tap(portfolios => this.portfoliosSubject.next(portfolios))
    );
  }

  getPortfolioById(id: string): Observable<Portfolio> {
    const cached = this.portfoliosSubject.value.find(p => p.id === id);
    if (cached) {
      this.currentPortfolioSubject.next(cached);
      return of(cached);
    }

    return this.getPortfolios().pipe(
      map(portfolios => {
        const portfolio = portfolios.find(p => p.id === id);
        if (!portfolio) {
          throw new Error('Portfolio not found');
        }
        return portfolio;
      }),
      tap(portfolio => this.currentPortfolioSubject.next(portfolio))
    );    
  }

  createPortfolio(data: { name: string }): Observable<Portfolio> {
    return this.http.post<any>(this.portfolioApiUrl, data).pipe(
      map(this.mapPortfolioFromApi),
      tap(portfolio => {
        const current = this.portfoliosSubject.value;
        this.portfoliosSubject.next([...current, portfolio]);
      })
    );
  }

  updatePortfolio(id: string, data: { name: string }): Observable<void> {
    return this.http.put<void>(`${this.portfolioApiUrl}/${id}`, data).pipe(
      tap(() => {
        const current = this.portfoliosSubject.value;
        const index = current.findIndex(p => p.id === id);
        if (index > -1) {
          current[index] = { ...current[index], name: data.name };
          this.portfoliosSubject.next([...current]);
        }
        if (this.currentPortfolioSubject.value?.id === id) {
          this.currentPortfolioSubject.next({ ...this.currentPortfolioSubject.value, name: data.name });
        }
      })
    );
  }

  deletePortfolio(id: string): Observable<void> {
    return this.http.delete<void>(`${this.portfolioApiUrl}/${id}`).pipe(
      tap(() => {
        const current = this.portfoliosSubject.value;
        this.portfoliosSubject.next(current.filter(p => p.id !== id));
      })
    );
  }

  getPortfolioMetrics(id: string): Observable<DashboardMetrics> {
    return this.http.get<DashboardMetrics>(`${this.portfolioApiUrl}/${id}/metrics`).pipe(
      tap(metrics => this.metricsSubject.next(metrics))
    );
  }

  /** Portfolio analytics summary (totals, P/L, allocation, diversification & volatility scores). */
  getPortfolioSummary(portfolioId: string): Observable<PortfolioSummary> {
    return this.http.get<any>(`${this.portfolioApiUrl}/${portfolioId}/summary`).pipe(
      map(raw => this.mapPortfolioSummaryFromApi(raw))
    );
  }

  createPortfolioSnapshot(portfolioId: string): Observable<PortfolioSnapshot> {
    return this.http.post<any>(`${this.portfolioApiUrl}/${portfolioId}/snapshots`, {}).pipe(
      map(raw => this.mapPortfolioSnapshotFromApi(raw, portfolioId))
    );
  }

  getPortfolioSnapshots(portfolioId: string): Observable<PortfolioSnapshot[]> {
    return this.http.get<any[]>(`${this.portfolioApiUrl}/${portfolioId}/snapshots`).pipe(
      map(rows => (rows || []).map(row => this.mapPortfolioSnapshotFromApi(row, portfolioId)))
    );
  }

  getPortfolioSnapshotById(portfolioId: string, snapshotId: string): Observable<PortfolioSnapshot> {
    return this.http.get<any>(`${this.portfolioApiUrl}/${portfolioId}/snapshots/${snapshotId}`).pipe(
      map(raw => this.mapPortfolioSnapshotFromApi(raw, portfolioId))
    );
  }

  getHoldings(portfolioId: string): Observable<Holding[]> {
    return this.http
      .get<any[]>(`${this.holdingsApiUrl}/${portfolioId}`)
      .pipe(
        map(holdings =>
          (holdings || []).map(h => ({
            id: h.id,
            portfolioId: h.portfolio_id ?? h.portfolioId,
            symbol: h.symbol,
            assetType: h.asset_type ?? h.assetType,
            quantity: Number(h.quantity),
            avgPrice: Number(h.avg_price ?? h.avgPrice),
            createdAt: h.created_at ? new Date(h.created_at) : h.createdAt,
            updatedAt: h.updated_at ? new Date(h.updated_at) : h.updatedAt
          }))
        )
      );
  }

  addHolding(data: {
    portfolio_id: string;
    symbol: string;
    asset_type: string;
    quantity: number;
    avg_price: number;
  }): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      this.holdingsApiUrl,
      data
    );
  }

  updateHolding(id: string, data: { quantity: number; avg_price: number }): Observable<{ message: string }> {
    return this.http.put<{ message: string }>(`${this.holdingsApiUrl}/${id}`, data);
  }

  deleteHolding(id: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(
      `${this.holdingsApiUrl}/${id}`
    );
  }

  private mapPortfolioFromApi = (p: any): Portfolio => ({
    id: p.id,
    userId: p.user_id ?? p.userId,
    name: p.name,
    createdAt: p.created_at ? new Date(p.created_at) : p.createdAt
  });

  private mapPortfolioSummaryFromApi = (s: any): PortfolioSummary => ({
    portfolioId: s.portfolio_id ?? s.portfolioId,
    portfolioName: s.portfolio_name ?? s.portfolioName,
    totalInvested: Number(s.total_invested ?? s.totalInvested ?? 0),
    totalPortfolioValue: Number(s.total_portfolio_value ?? s.totalPortfolioValue ?? 0),
    totalProfitLoss: Number(s.total_profit_loss ?? s.totalProfitLoss ?? 0),
    profitLossPercentage: Number(s.profit_loss_percentage ?? s.profitLossPercentage ?? 0),
    diversificationScore: Number(s.diversification_score ?? s.diversificationScore ?? 0),
    volatilityScore: Number(s.volatility_score ?? s.volatilityScore ?? 0),
    assetAllocation: (s.asset_allocation ?? s.assetAllocation ?? []).map((row: any) =>
      this.mapAssetAllocationRowFromApi(row)
    )
  });

  private mapAssetAllocationRowFromApi = (row: any): AssetAllocationRow => ({
    symbol: row.symbol,
    assetType: row.asset_type ?? row.assetType ?? '',
    costBasis: Number(row.cost_basis ?? row.costBasis ?? 0),
    currentPrice: Number(row.current_price ?? row.currentPrice ?? 0),
    value: Number(row.value ?? 0),
    percent: Number(row.percent ?? 0)
  });

  private mapPortfolioSnapshotFromApi(raw: any, portfolioId: string): PortfolioSnapshot {
    let summaryRaw: any = raw.summary ?? raw.portfolio_summary ?? raw.summary_json ?? raw;
    if (typeof summaryRaw === 'string') {
      try {
        summaryRaw = JSON.parse(summaryRaw);
      } catch {
        summaryRaw = {};
      }
    }
    return {
      id: String(raw.id ?? raw.snapshot_id ?? raw.snapshotId ?? ''),
      portfolioId: String(raw.portfolio_id ?? raw.portfolioId ?? portfolioId),
      portfolioName: raw.portfolio_name ?? raw.portfolioName ?? summaryRaw?.portfolio_name ?? summaryRaw?.portfolioName,
      createdAt: new Date(raw.created_at ?? raw.createdAt ?? new Date().toISOString()),
      summary: this.mapPortfolioSummaryFromApi(summaryRaw)
    };
  }
}
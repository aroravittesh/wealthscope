import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { BehaviorSubject, Observable } from 'rxjs';
import { map, tap } from 'rxjs/operators';

import { DashboardMetrics, Holding, Portfolio } from '../models';

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
    return this.http.get<Portfolio[]>(this.portfolioApiUrl).pipe(
      tap(portfolios => this.portfoliosSubject.next(portfolios))
    );
  }

  // ✅ FIXED HERE
  getPortfolioById(id: string): Observable<Portfolio> {
    return this.http.get<Portfolio>(`${this.portfolioApiUrl}/${id}`).pipe(
      tap(portfolio => this.currentPortfolioSubject.next(portfolio))
    );
  }

  createPortfolio(data: { name: string; description: string }): Observable<Portfolio> {
    return this.http.post<Portfolio>(this.portfolioApiUrl, data).pipe(
      tap(portfolio => {
        const current = this.portfoliosSubject.value;
        this.portfoliosSubject.next([...current, portfolio]);
      })
    );
  }

  updatePortfolio(id: string, data: { name: string; description: string }): Observable<Portfolio> {
    return this.http.put<Portfolio>(`${this.portfolioApiUrl}/${id}`, data).pipe(
      tap(portfolio => {
        const current = this.portfoliosSubject.value;
        const index = current.findIndex(p => p.id === id);
        if (index > -1) {
          current[index] = portfolio;
          this.portfoliosSubject.next([...current]);
        }
        if (this.currentPortfolioSubject.value?.id === id) {
          this.currentPortfolioSubject.next(portfolio);
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

  // ✅ HOLDINGS FETCH
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

  // ✅ ADD / UPDATE HOLDING → hits POST /holdings
  addOrUpdateHolding(data: {
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

  // ✅ DELETE HOLDING → DELETE /holdings/:id
  deleteHolding(id: string): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(
      `${this.holdingsApiUrl}/${id}`
    );
  }
}
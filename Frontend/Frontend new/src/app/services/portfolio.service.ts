import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { tap, map, delay, catchError } from 'rxjs/operators';
import { Portfolio, Holding, DashboardMetrics } from '../models';

@Injectable({
  providedIn: 'root'
})
export class PortfolioService {
  private apiUrl = `http://${window.location.hostname}:8080/api/portfolios`;
  private portfoliosSubject = new BehaviorSubject<Portfolio[]>([]);
  public portfolios$ = this.portfoliosSubject.asObservable();

  private currentPortfolioSubject = new BehaviorSubject<Portfolio | null>(null);
  public currentPortfolio$ = this.currentPortfolioSubject.asObservable();

  private metricsSubject = new BehaviorSubject<DashboardMetrics | null>(null);
  public metrics$ = this.metricsSubject.asObservable();

  constructor(private http: HttpClient) {}

  getPortfolios(): Observable<Portfolio[]> {
    const mockPortfolios = [
      { id: '1', name: 'Tech Growth', description: 'High growth tech sector stocks', totalValue: 4500.50, profitLossPercentage: 12.5 },
      { id: '2', name: 'Dividend Yield', description: 'Stable dividend paying stocks', totalValue: 8200.00, profitLossPercentage: 4.2 }
    ] as any[];

    return of(mockPortfolios).pipe(
      delay(800), // Simulate network delay
      tap(portfolios => this.portfoliosSubject.next(portfolios))
    );
  }

  getPortfolioById(id: string): Observable<Portfolio> {
    return this.http.get<Portfolio>(`${this.apiUrl}/${id}`).pipe(
      tap(portfolio => this.currentPortfolioSubject.next(portfolio))
    );
  }

  createPortfolio(data: { name: string; description: string }): Observable<Portfolio> {
    return this.http.post<Portfolio>(this.apiUrl, data).pipe(
      tap(portfolio => {
        const current = this.portfoliosSubject.value;
        this.portfoliosSubject.next([...current, portfolio]);
      })
    );
  }

  updatePortfolio(id: string, data: { name: string; description: string }): Observable<Portfolio> {
    return this.http.put<Portfolio>(`${this.apiUrl}/${id}`, data).pipe(
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
    return this.http.delete<void>(`${this.apiUrl}/${id}`).pipe(
      tap(() => {
        const current = this.portfoliosSubject.value;
        this.portfoliosSubject.next(current.filter(p => p.id !== id));
      })
    );
  }

  getPortfolioMetrics(id: string): Observable<DashboardMetrics> {
    return this.http.get<DashboardMetrics>(`${this.apiUrl}/${id}/metrics`).pipe(
      tap(metrics => this.metricsSubject.next(metrics)),
      catchError((err: any) => {
        console.warn('Backend metrics failed, using mock data...', err);
        const mockMetrics: DashboardMetrics = {
          totalPortfolioValue: 125430.50,
          totalInvested: 100000.00,
          totalProfitLoss: 25430.50,
          profitLossPercentage: 25.4,
          assetsCount: 12,
          portfoliosCount: 1,
          topPerformers: [],
          allocationData: { 'Stocks': 60, 'Bonds': 30, 'Cash': 10 }
        };
        return of(mockMetrics).pipe(
          delay(1500), // Simulate network delay so the fancy loader renders
          tap(metrics => this.metricsSubject.next(metrics))
        );
      })
    );
  }

  getHoldings(portfolioId: string): Observable<Holding[]> {
    return this.http.get<Holding[]>(`${this.apiUrl}/${portfolioId}/holdings`);
  }
}

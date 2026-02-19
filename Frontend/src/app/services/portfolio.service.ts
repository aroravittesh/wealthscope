import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { tap, map } from 'rxjs/operators';
import { Portfolio, Holding, DashboardMetrics } from '../models';

@Injectable({
  providedIn: 'root'
})
export class PortfolioService {
  private apiUrl = 'http://localhost:8080/api/portfolios';
  private portfoliosSubject = new BehaviorSubject<Portfolio[]>([]);
  public portfolios$ = this.portfoliosSubject.asObservable();

  private currentPortfolioSubject = new BehaviorSubject<Portfolio | null>(null);
  public currentPortfolio$ = this.currentPortfolioSubject.asObservable();

  private metricsSubject = new BehaviorSubject<DashboardMetrics | null>(null);
  public metrics$ = this.metricsSubject.asObservable();

  constructor(private http: HttpClient) {}

  getPortfolios(): Observable<Portfolio[]> {
    return this.http.get<Portfolio[]>(this.apiUrl).pipe(
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
      tap(metrics => this.metricsSubject.next(metrics))
    );
  }

  getHoldings(portfolioId: string): Observable<Holding[]> {
    return this.http.get<Holding[]>(`${this.apiUrl}/${portfolioId}/holdings`);
  }
}

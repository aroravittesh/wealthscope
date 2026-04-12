import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface PortfolioItem {
  stock: string;
}

export interface RecommendationRequest {
  user_portfolio: PortfolioItem[];
  risk?: string;
  horizon?: string;
  top_n?: number;
}

export interface RecommendationResponse {
  stock: string;
  score: number;
  decision: string;
  predicted_return: number;
  sharpe: number;
  volatility: number;
  momentum: number;
  reason: string;
}

@Injectable({ providedIn: 'root' })
export class StockRecommendationService {
  private apiUrl = `http://${window.location.hostname}:8000/recommend`; // Updated to point to ML service port 8000

  constructor(private http: HttpClient) {}

  getRecommendations(request: RecommendationRequest): Observable<RecommendationResponse[]> {
    return this.http.post<RecommendationResponse[]>(this.apiUrl, request);
  }
}

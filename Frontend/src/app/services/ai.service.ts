import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import {
  AIRecommendRequest,
  AIRecommendResponse
} from '../models';

@Injectable({
  providedIn: 'root'
})
export class AiService {
  private aiApiUrl = `${environment.apiUrl}/ai`;

  constructor(private http: HttpClient) {}

  recommend(payload: AIRecommendRequest): Observable<AIRecommendResponse[]> {
    const requestBody = {
      user_portfolio: (payload.userPortfolio || []).map(item => ({ stock: item.stock })),
      risk: payload.risk ?? 'medium',
      horizon: payload.horizon ?? 'long',
      top_n: payload.topN ?? 5
    };
    return this.http.post<any[]>(`${this.aiApiUrl}/recommend`, requestBody).pipe(
      // Map snake_case API response to frontend camelCase model
      // without changing your Python service contract.
      map(rows =>
        (rows || []).map(row => ({
          stock: row.stock,
          score: Number(row.score ?? 0),
          decision: row.decision ?? 'HOLD',
          predictedReturn: Number(row.predicted_return ?? 0),
          sharpe: Number(row.sharpe ?? 0),
          volatility: Number(row.volatility ?? 0),
          momentum: Number(row.momentum ?? 0),
          reason: row.reason ?? ''
        }))
      )
    );
  }
}

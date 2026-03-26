import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { StockRecommendationService, RecommendationRequest, RecommendationResponse } from './services/stock-recommendation.service';

@Component({
  selector: 'app-ai-recommendations',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './ai-recommendations.component.html',
  styleUrls: ['./ai-recommendations.component.scss']
})
export class AiRecommendationsComponent {
  recommendations: RecommendationResponse[] = [];
  loading = false;
  error: string | null = null;

  portfolioInput = 'AAPL,TSLA';
  risk = 'medium';
  horizon = 'long';

  constructor(private recService: StockRecommendationService) {}

  getRecommendations() {
    this.loading = true;
    this.error = null;
    // Parse portfolio input into array of {stock: string}
    const user_portfolio = this.portfolioInput.split(',').map(s => ({ stock: s.trim().toUpperCase() })).filter(s => s.stock);
    const req: RecommendationRequest = {
      user_portfolio,
      risk: this.risk,
      horizon: this.horizon,
      top_n: 5
    };
    this.recService.getRecommendations(req).subscribe({
      next: (data: RecommendationResponse[]) => {
        this.recommendations = data;
        this.loading = false;
      },
      error: (err: any) => {
        this.error = 'Failed to fetch recommendations.';
        this.loading = false;
      }
    });
  }
}

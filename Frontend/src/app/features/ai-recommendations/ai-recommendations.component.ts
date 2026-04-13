import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AiService } from '../../services/ai.service';
import { AIRecommendRequest, AIRecommendResponse } from '../../models';

@Component({
  selector: 'app-ai-recommendations',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './ai-recommendations.component.html',
  styleUrl: './ai-recommendations.component.scss'
})
export class AiRecommendationsComponent {
  recommendations: AIRecommendResponse[] = [];
  loading = false;
  error: string | null = null;

  portfolioInput = 'AAPL,TSLA';
  risk = 'medium';
  horizon = 'long';

  constructor(private aiService: AiService) {}

  getRecommendations(): void {
    this.loading = true;
    this.error = null;

    const userPortfolio = this.portfolioInput
      .split(',')
      .map(s => ({ stock: s.trim().toUpperCase() }))
      .filter(s => !!s.stock);

    const req: AIRecommendRequest = {
      userPortfolio,
      risk: this.risk,
      horizon: this.horizon,
      topN: 5
    };

    this.aiService.recommend(req).subscribe({
      next: data => {
        this.recommendations = data;
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to fetch recommendations.';
        this.loading = false;
      }
    });
  }
}

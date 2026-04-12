
import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { AiRecommendationsComponent } from './ai-recommendations.component';

@NgModule({
  imports: [CommonModule, FormsModule, HttpClientModule, AiRecommendationsComponent],
  exports: [AiRecommendationsComponent]
})
export class AiRecommendationsModule {}

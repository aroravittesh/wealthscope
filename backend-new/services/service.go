package services

import (
	"context"
	"stock-backend/models"
)

// StockService defines business logic for stock recommendations
// FUTURE: This will call the FastAPI ML microservice for predictions

type StockService interface {
	GetRecommendation(ctx context.Context, req models.StockRecommendationRequest) (*models.StockRecommendationResponse, error)
}

type stockService struct{}

func NewStockService() StockService {
	return &stockService{}
}

// GetRecommendation returns a placeholder response for now
func (s *stockService) GetRecommendation(ctx context.Context, req models.StockRecommendationRequest) (*models.StockRecommendationResponse, error) {
	// TODO: Integrate with FastAPI ML microservice here
	return &models.StockRecommendationResponse{
		RecommendationScore: 0.5,
		Confidence:          0.0,
		Status:              "pending_ml_service",
	}, nil
}

package repository

import "wealthscope-backend/internal/models"

type PortfolioSnapshotRepository interface {
	Create(s *models.PortfolioSnapshot) error
	ListByPortfolio(portfolioID string) ([]models.PortfolioSnapshot, error)
}

package services

import (
	"strings"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type HoldingService struct {
	Repo repository.HoldingRepository
}


func (s *HoldingService) AddHolding(
	portfolioID string,
	symbol string,
	assetType string,
	quantity float64,
	avgPrice float64,
) error {

	h := &models.Holding{
		PortfolioID: portfolioID,
		Symbol:      strings.ToUpper(symbol),
		AssetType:   assetType,
		Quantity:    quantity,
		AvgPrice:    avgPrice,
	}

	return s.Repo.CreateOrUpdate(h)
}

func (s *HoldingService) GetHoldings(portfolioID string) ([]models.Holding, error) {
	return s.Repo.GetByPortfolio(portfolioID)
}

func (s *HoldingService) DeleteHolding(id string) error {
	return s.Repo.Delete(id)
}


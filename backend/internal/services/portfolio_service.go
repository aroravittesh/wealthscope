package services

import (
	"errors"
	"math"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type PortfolioService struct {
	PortfolioRepo repository.PortfolioRepository
	HoldingRepo   repository.HoldingRepository
}

func (s *PortfolioService) Create(userID string, name string) (*models.Portfolio, error) {
	if name == "" {
		return nil, errors.New("portfolio name required")
	}

	p := &models.Portfolio{
		UserID: userID,
		Name:   name,
	}

	err := s.PortfolioRepo.Create(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PortfolioService) GetUserPortfolios(userID string) ([]models.Portfolio, error) {
	return s.PortfolioRepo.GetByUser(userID)
}

func (s *PortfolioService) Rename(userID, portfolioID, name string) error {
	p, err := s.PortfolioRepo.GetByID(portfolioID)
	if err != nil {
		return err
	}

	if p.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.PortfolioRepo.UpdateName(portfolioID, name)
}

func (s *PortfolioService) Delete(userID, portfolioID string) error {
	p, err := s.PortfolioRepo.GetByID(portfolioID)
	if err != nil {
		return err
	}

	if p.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.PortfolioRepo.Delete(portfolioID)
}

// GetPortfolioSummary aggregates holdings for analytics. Requires HoldingRepo.
// Slice 1: total_portfolio_value equals cost basis (quantity × avg_price); P/L is zero.
func (s *PortfolioService) GetPortfolioSummary(userID, portfolioID string) (*models.PortfolioSummary, error) {
	if s.HoldingRepo == nil {
		return nil, errors.New("holding repository not configured")
	}

	p, err := s.PortfolioRepo.GetByID(portfolioID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	holdings, err := s.HoldingRepo.GetByPortfolio(portfolioID)
	if err != nil {
		return nil, err
	}

	var total float64
	rows := make([]models.AssetAllocationRow, 0, len(holdings))
	for _, h := range holdings {
		v := h.Quantity * h.AvgPrice
		total += v
		rows = append(rows, models.AssetAllocationRow{
			Symbol:    h.Symbol,
			AssetType: h.AssetType,
			Value:     v,
			Percent:   0,
		})
	}

	if total > 0 {
		for i := range rows {
			rows[i].Percent = math.Round((rows[i].Value/total)*10000) / 100
		}
	}

	return &models.PortfolioSummary{
		PortfolioID:          p.ID,
		PortfolioName:        p.Name,
		TotalInvested:        total,
		TotalPortfolioValue:  total,
		TotalProfitLoss:      0,
		ProfitLossPercentage: 0,
		AssetAllocation:      rows,
	}, nil
}

package services

import (
	"errors"
	"math"

	"wealthscope-backend/internal/market"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type PortfolioService struct {
	PortfolioRepo repository.PortfolioRepository
	HoldingRepo   repository.HoldingRepository
	Prices        market.PriceProvider
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
// Totals use mark-to-estimate prices from Prices (defaults to cost basis if Prices is nil).
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

	prices := s.Prices
	if prices == nil {
		prices = market.Passthrough{}
	}

	var totalInvested float64
	type rowAgg struct {
		row models.AssetAllocationRow
	}
	aggs := make([]rowAgg, 0, len(holdings))

	for _, h := range holdings {
		cost := h.Quantity * h.AvgPrice
		totalInvested += cost
		unit := prices.UnitPrice(h.Symbol, h.AvgPrice)
		mkt := h.Quantity * unit
		aggs = append(aggs, rowAgg{
			row: models.AssetAllocationRow{
				Symbol:       h.Symbol,
				AssetType:    h.AssetType,
				CostBasis:    cost,
				CurrentPrice: unit,
				Value:        mkt,
				Percent:      0,
			},
		})
	}

	var totalMkt float64
	for _, a := range aggs {
		totalMkt += a.row.Value
	}

	if totalMkt > 0 {
		for i := range aggs {
			aggs[i].row.Percent = math.Round((aggs[i].row.Value/totalMkt)*10000) / 100
		}
	}

	rows := make([]models.AssetAllocationRow, len(aggs))
	for i := range aggs {
		rows[i] = aggs[i].row
	}

	pnl := totalMkt - totalInvested
	var pnlPct float64
	if totalInvested > 0 {
		pnlPct = math.Round((pnl/totalInvested)*10000) / 100
	}

	return &models.PortfolioSummary{
		PortfolioID:          p.ID,
		PortfolioName:        p.Name,
		TotalInvested:        totalInvested,
		TotalPortfolioValue:  totalMkt,
		TotalProfitLoss:      pnl,
		ProfitLossPercentage: pnlPct,
		AssetAllocation:      rows,
	}, nil
}

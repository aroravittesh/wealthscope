package services

import (
	"errors"
	"strings"

	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type HoldingService struct {
	Repo          repository.HoldingRepository
	PortfolioRepo repository.PortfolioRepository
}

func (s *HoldingService) AddHolding(
	userID string,
	portfolioID string,
	symbol string,
	assetType string,
	quantity float64,
	avgPrice float64,
) error {
	if err := s.ensurePortfolioOwnership(userID, portfolioID); err != nil {
		return err
	}

	h := &models.Holding{
		PortfolioID: portfolioID,
		Symbol:      strings.ToUpper(symbol),
		AssetType:   assetType,
		Quantity:    quantity,
		AvgPrice:    avgPrice,
	}

	return s.Repo.CreateOrUpdate(h)
}

func (s *HoldingService) GetHoldings(userID, portfolioID string) ([]models.Holding, error) {
	if err := s.ensurePortfolioOwnership(userID, portfolioID); err != nil {
		return nil, err
	}
	return s.Repo.GetByPortfolio(portfolioID)
}

func (s *HoldingService) DeleteHolding(userID, id string) error {
	holding, err := s.Repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.ensurePortfolioOwnership(userID, holding.PortfolioID); err != nil {
		return err
	}
	return s.Repo.Delete(id)
}

func (s *HoldingService) UpdateHolding(userID, id string, quantity float64, avgPrice float64) error {
	holding, err := s.Repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.ensurePortfolioOwnership(userID, holding.PortfolioID); err != nil {
		return err
	}
	return s.Repo.UpdateByID(id, quantity, avgPrice)
}

func (s *HoldingService) ensurePortfolioOwnership(userID, portfolioID string) error {
	p, err := s.PortfolioRepo.GetByID(portfolioID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return errors.New("unauthorized")
	}
	return nil
}

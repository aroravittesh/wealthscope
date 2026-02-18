package services

import (
	"errors"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type PortfolioService struct {
	PortfolioRepo repository.PortfolioRepository
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

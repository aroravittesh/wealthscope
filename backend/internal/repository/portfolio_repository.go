package repository

import "wealthscope-backend/internal/models"

type PortfolioRepository interface {
	Create(portfolio *models.Portfolio) error
	GetByUser(userID string) ([]models.Portfolio, error)
	GetByID(id string) (*models.Portfolio, error)
	UpdateName(id string, name string) error
	Delete(id string) error
}

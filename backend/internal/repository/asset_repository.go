package repository

import "wealthscope-backend/internal/models"

type AssetRepository interface {
	List() ([]models.Asset, error)
	Create(a *models.Asset) error
	Update(id string, symbol string, name string, assetType string) error
	Delete(id string) error
	FindByID(id string) (*models.Asset, error)
}

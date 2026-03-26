package repository

import "wealthscope-backend/internal/models"

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	UpdatePassword(userID string, passwordHash string) error
	UpdateRiskPreference(userID string, riskPreference string) error
}

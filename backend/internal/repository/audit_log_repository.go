package repository

import "wealthscope-backend/internal/models"

type AuditLogRepository interface {
	Create(log *models.AuditLog) error
}

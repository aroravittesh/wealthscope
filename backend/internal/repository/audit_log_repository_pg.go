package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"wealthscope-backend/internal/models"
)

type AuditLogRepositoryPG struct {
	DB *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepositoryPG {
	return &AuditLogRepositoryPG{DB: db}
}

func (r *AuditLogRepositoryPG) Create(log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	_, err := r.DB.Exec(`
		INSERT INTO audit_logs (id, actor_user_id, action, entity_type, entity_id, before_json, after_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		log.ID,
		log.ActorUserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.BeforeJSON,
		log.AfterJSON,
		log.CreatedAt,
	)
	return err
}

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

func (r *AuditLogRepositoryPG) List(limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.Query(`
		SELECT id, COALESCE(actor_user_id::text, ''), action, entity_type, COALESCE(entity_id::text, ''), COALESCE(before_json, ''), COALESCE(after_json, ''), created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.AuditLog, 0, limit)
	for rows.Next() {
		var a models.AuditLog
		if err := rows.Scan(
			&a.ID,
			&a.ActorUserID,
			&a.Action,
			&a.EntityType,
			&a.EntityID,
			&a.BeforeJSON,
			&a.AfterJSON,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

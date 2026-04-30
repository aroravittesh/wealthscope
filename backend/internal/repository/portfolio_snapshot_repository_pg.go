package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"wealthscope-backend/internal/models"
)

type PortfolioSnapshotRepositoryPG struct {
	DB *sql.DB
}

func NewPortfolioSnapshotRepository(db *sql.DB) *PortfolioSnapshotRepositoryPG {
	return &PortfolioSnapshotRepositoryPG{DB: db}
}

func (r *PortfolioSnapshotRepositoryPG) Create(s *models.PortfolioSnapshot) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	_, err := r.DB.Exec(`
		INSERT INTO portfolio_snapshots (id, portfolio_id, user_id, summary_json, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, s.ID, s.PortfolioID, s.UserID, s.SummaryJSON, time.Now())
	return err
}

func (r *PortfolioSnapshotRepositoryPG) ListByPortfolio(portfolioID string) ([]models.PortfolioSnapshot, error) {
	rows, err := r.DB.Query(`
		SELECT id, portfolio_id, user_id, summary_json, created_at
		FROM portfolio_snapshots
		WHERE portfolio_id = $1
		ORDER BY created_at DESC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.PortfolioSnapshot
	for rows.Next() {
		var s models.PortfolioSnapshot
		if err := rows.Scan(&s.ID, &s.PortfolioID, &s.UserID, &s.SummaryJSON, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

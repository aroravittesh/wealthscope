package repository

import (
	"database/sql"
	"wealthscope-backend/internal/models"
)

type PortfolioRepositoryPG struct {
	DB *sql.DB
}

func NewPortfolioRepository(db *sql.DB) *PortfolioRepositoryPG {
	return &PortfolioRepositoryPG{DB: db}
}

func (r *PortfolioRepositoryPG) Create(p *models.Portfolio) error {
	query := `
		INSERT INTO portfolios (user_id, name)
		VALUES ($1, $2)
		RETURNING id, created_at;
	`
	return r.DB.QueryRow(query, p.UserID, p.Name).
		Scan(&p.ID, &p.CreatedAt)
}

func (r *PortfolioRepositoryPG) GetByUser(userID string) ([]models.Portfolio, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM portfolios
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []models.Portfolio

	for rows.Next() {
		var p models.Portfolio
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		portfolios = append(portfolios, p)
	}

	return portfolios, nil
}

func (r *PortfolioRepositoryPG) GetByID(id string) (*models.Portfolio, error) {
	query := `
		SELECT id, user_id, name, created_at
		FROM portfolios
		WHERE id = $1;
	`

	var p models.Portfolio
	err := r.DB.QueryRow(query, id).
		Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PortfolioRepositoryPG) UpdateName(id string, name string) error {
	_, err := r.DB.Exec(
		`UPDATE portfolios SET name = $1 WHERE id = $2`,
		name, id,
	)
	return err
}

func (r *PortfolioRepositoryPG) Delete(id string) error {
	_, err := r.DB.Exec(
		`DELETE FROM portfolios WHERE id = $1`,
		id,
	)
	return err
}

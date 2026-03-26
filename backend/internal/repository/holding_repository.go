package repository

import (
	"database/sql"
	"time"

	"wealthscope-backend/internal/models"

	"github.com/google/uuid"
)

type HoldingRepository interface {
	CreateOrUpdate(h *models.Holding) error
	GetByPortfolio(portfolioID string) ([]models.Holding, error)
	Delete(id string) error
	FindBySymbol(portfolioID, symbol string) (*models.Holding, error)
}

type HoldingRepositoryPG struct {
	DB *sql.DB
}

func NewHoldingRepository(db *sql.DB) *HoldingRepositoryPG {
	return &HoldingRepositoryPG{DB: db}
}

func (r *HoldingRepositoryPG) CreateOrUpdate(h *models.Holding) error {

	existing, _ := r.FindBySymbol(h.PortfolioID, h.Symbol)

	if existing != nil {
		// update quantity
		_, err := r.DB.Exec(`
			UPDATE holdings
			SET quantity = $1,
			    avg_price = $2,
			    updated_at = $3
			WHERE id = $4
		`, h.Quantity, h.AvgPrice, time.Now(), existing.ID)

		return err
	}

	h.ID = uuid.New().String()

	_, err := r.DB.Exec(`
		INSERT INTO holdings (id, portfolio_id, symbol, asset_type, quantity, avg_price, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, h.ID, h.PortfolioID, h.Symbol, h.AssetType, h.Quantity, h.AvgPrice, time.Now(), time.Now())

	return err
}

func (r *HoldingRepositoryPG) FindBySymbol(portfolioID, symbol string) (*models.Holding, error) {

	row := r.DB.QueryRow(`
		SELECT id, portfolio_id, symbol, asset_type, quantity, avg_price, created_at, updated_at
		FROM holdings
		WHERE portfolio_id=$1 AND symbol=$2
	`, portfolioID, symbol)

	var h models.Holding

	err := row.Scan(
		&h.ID, &h.PortfolioID, &h.Symbol, &h.AssetType,
		&h.Quantity, &h.AvgPrice, &h.CreatedAt, &h.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (r *HoldingRepositoryPG) GetByPortfolio(portfolioID string) ([]models.Holding, error) {

	rows, err := r.DB.Query(`
		SELECT id, portfolio_id, symbol, asset_type, quantity, avg_price, created_at, updated_at
		FROM holdings
		WHERE portfolio_id=$1
	`, portfolioID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []models.Holding

	for rows.Next() {
		var h models.Holding

		err := rows.Scan(
			&h.ID, &h.PortfolioID, &h.Symbol, &h.AssetType,
			&h.Quantity, &h.AvgPrice, &h.CreatedAt, &h.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		holdings = append(holdings, h)
	}

	return holdings, nil
}

func (r *HoldingRepositoryPG) Delete(id string) error {
	_, err := r.DB.Exec("DELETE FROM holdings WHERE id=$1", id)
	return err
}
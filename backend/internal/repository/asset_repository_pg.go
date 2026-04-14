package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"wealthscope-backend/internal/models"
)

type AssetRepositoryPG struct {
	DB *sql.DB
}

func NewAssetRepository(db *sql.DB) *AssetRepositoryPG {
	return &AssetRepositoryPG{DB: db}
}

func (r *AssetRepositoryPG) List() ([]models.Asset, error) {
	rows, err := r.DB.Query(`
		SELECT id, symbol, COALESCE(name, ''), asset_type, created_at
		FROM assets
		ORDER BY symbol
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Asset
	for rows.Next() {
		var a models.Asset
		if err := rows.Scan(&a.ID, &a.Symbol, &a.Name, &a.AssetType, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AssetRepositoryPG) Create(a *models.Asset) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	_, err := r.DB.Exec(`
		INSERT INTO assets (id, symbol, name, asset_type, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5)
	`, a.ID, a.Symbol, a.Name, a.AssetType, time.Now())
	return err
}

func (r *AssetRepositoryPG) Update(id string, symbol string, name string, assetType string) error {
	res, err := r.DB.Exec(`
		UPDATE assets
		SET symbol = $1, name = NULLIF($2, ''), asset_type = $3
		WHERE id = $4
	`, symbol, name, assetType, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("asset not found")
	}
	return nil
}

func (r *AssetRepositoryPG) Delete(id string) error {
	res, err := r.DB.Exec(`DELETE FROM assets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("asset not found")
	}
	return nil
}

func (r *AssetRepositoryPG) FindByID(id string) (*models.Asset, error) {
	row := r.DB.QueryRow(`
		SELECT id, symbol, COALESCE(name, ''), asset_type, created_at
		FROM assets WHERE id = $1
	`, id)
	var a models.Asset
	err := row.Scan(&a.ID, &a.Symbol, &a.Name, &a.AssetType, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("asset not found")
	}
	return &a, err
}

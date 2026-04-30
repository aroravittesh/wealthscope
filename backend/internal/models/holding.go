package models

import "time"

type Holding struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolio_id"`
	Symbol      string    `json:"symbol"`
	AssetType   string    `json:"asset_type"`
	Quantity    float64   `json:"quantity"`
	AvgPrice    float64   `json:"avg_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

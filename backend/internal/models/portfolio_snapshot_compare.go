package models

import "time"

type SnapshotDelta struct {
	Absolute float64 `json:"absolute"`
	Percent  float64 `json:"percent"`
}

type PortfolioSnapshotCompareResponse struct {
	PortfolioID string    `json:"portfolio_id"`
	FromID      string    `json:"from_id"`
	ToID        string    `json:"to_id"`
	FromAt      time.Time `json:"from_at"`
	ToAt        time.Time `json:"to_at"`

	TotalValueDelta      SnapshotDelta `json:"total_value_delta"`
	TotalInvestedDelta   SnapshotDelta `json:"total_invested_delta"`
	ProfitLossDelta      SnapshotDelta `json:"profit_loss_delta"`
	DiversificationDelta SnapshotDelta `json:"diversification_delta"`
	VolatilityDelta      SnapshotDelta `json:"volatility_delta"`
}

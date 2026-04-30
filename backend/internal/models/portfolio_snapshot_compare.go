package models

import "time"

type SnapshotDelta struct {
	Absolute float64 `json:"absolute"`
	Percent  float64 `json:"percent"`
}

type AllocationDriftRow struct {
	Symbol       string  `json:"symbol"`
	FromPercent  float64 `json:"from_percent"`
	ToPercent    float64 `json:"to_percent"`
	DeltaPercent float64 `json:"delta_percent"`
	FromValue    float64 `json:"from_value"`
	ToValue      float64 `json:"to_value"`
	DeltaValue   float64 `json:"delta_value"`
}

type PortfolioSnapshotCompareResponse struct {
	PortfolioID string    `json:"portfolio_id"`
	FromID      string    `json:"from_id"`
	ToID        string    `json:"to_id"`
	FromAt      time.Time `json:"from_at"`
	ToAt        time.Time `json:"to_at"`

	TotalValueDelta      SnapshotDelta        `json:"total_value_delta"`
	TotalInvestedDelta   SnapshotDelta        `json:"total_invested_delta"`
	ProfitLossDelta      SnapshotDelta        `json:"profit_loss_delta"`
	DiversificationDelta SnapshotDelta        `json:"diversification_delta"`
	VolatilityDelta      SnapshotDelta        `json:"volatility_delta"`
	AllocationDrift      []AllocationDriftRow `json:"allocation_drift"`
}

type PortfolioSnapshotTrendPoint struct {
	SnapshotID          string    `json:"snapshot_id"`
	CreatedAt           time.Time `json:"created_at"`
	TotalPortfolioValue float64   `json:"total_portfolio_value"`
	TotalInvested       float64   `json:"total_invested"`
	TotalProfitLoss     float64   `json:"total_profit_loss"`
	Diversification     float64   `json:"diversification"`
	Volatility          float64   `json:"volatility"`
}

type PortfolioSnapshotTrendResponse struct {
	PortfolioID string                        `json:"portfolio_id"`
	Points      []PortfolioSnapshotTrendPoint `json:"points"`
}

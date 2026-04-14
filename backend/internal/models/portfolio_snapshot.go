package models

import "time"

// PortfolioSnapshot stores a point-in-time JSON snapshot of portfolio summary analytics.
type PortfolioSnapshot struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolio_id"`
	UserID      string    `json:"user_id"`
	SummaryJSON string    `json:"summary_json"`
	CreatedAt   time.Time `json:"created_at"`
}

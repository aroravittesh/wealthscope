package models

// PortfolioSummary is holdings-based analytics for a portfolio.
// Slice 1: market value and P/L match cost basis until live pricing (slice 2).
type PortfolioSummary struct {
	PortfolioID          string               `json:"portfolio_id"`
	PortfolioName        string               `json:"portfolio_name"`
	TotalInvested        float64              `json:"total_invested"`
	TotalPortfolioValue  float64              `json:"total_portfolio_value"`
	TotalProfitLoss      float64              `json:"total_profit_loss"`
	ProfitLossPercentage float64              `json:"profit_loss_percentage"`
	AssetAllocation      []AssetAllocationRow `json:"asset_allocation"`
}

type AssetAllocationRow struct {
	Symbol    string  `json:"symbol"`
	AssetType string  `json:"asset_type"`
	Value     float64 `json:"value"`
	Percent   float64 `json:"percent"`
}

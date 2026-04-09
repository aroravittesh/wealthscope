package models

// PortfolioSummary is holdings-based analytics for a portfolio.
// Total portfolio value and per-row values use mark-to-estimate unit prices (Slice 2).
type PortfolioSummary struct {
	PortfolioID          string               `json:"portfolio_id"`
	PortfolioName        string               `json:"portfolio_name"`
	TotalInvested        float64              `json:"total_invested"`
	TotalPortfolioValue  float64              `json:"total_portfolio_value"`
	TotalProfitLoss      float64              `json:"total_profit_loss"`
	ProfitLossPercentage float64              `json:"profit_loss_percentage"`
	DiversificationScore float64              `json:"diversification_score"`
	VolatilityScore      float64              `json:"volatility_score"`
	AssetAllocation      []AssetAllocationRow `json:"asset_allocation"`
}

// AssetAllocationRow: Value and Percent use market value (qty × current unit price).
type AssetAllocationRow struct {
	Symbol       string  `json:"symbol"`
	AssetType    string  `json:"asset_type"`
	CostBasis    float64 `json:"cost_basis"`
	CurrentPrice float64 `json:"current_price"`
	Value        float64 `json:"value"`
	Percent      float64 `json:"percent"`
}

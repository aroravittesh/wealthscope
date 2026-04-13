package models

type AIRecommendationPortfolioItem struct {
	Stock string `json:"stock"`
}

type AIRecommendRequest struct {
	UserPortfolio []AIRecommendationPortfolioItem `json:"user_portfolio"`
	Risk          string                          `json:"risk,omitempty"`
	Horizon       string                          `json:"horizon,omitempty"`
	TopN          int                             `json:"top_n,omitempty"`
}

type AIRecommendResponse struct {
	Stock           string  `json:"stock"`
	Score           float64 `json:"score"`
	Decision        string  `json:"decision"`
	PredictedReturn float64 `json:"predicted_return"`
	Sharpe          float64 `json:"sharpe"`
	Volatility      float64 `json:"volatility"`
	Momentum        float64 `json:"momentum"`
	Reason          string  `json:"reason"`
}

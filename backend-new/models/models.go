package models

// HealthCheckResponse represents the response for health check endpoint
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// StockRecommendationRequest represents a request for stock recommendation
// FUTURE: This will be used to call the FastAPI ML microservice
type StockRecommendationRequest struct {
	MarketHistory    [][]float64 `json:"market_history"`    // 60x5 array of market data
	UserProfile      []float64   `json:"user_profile"`      // 2 elements: risk tolerance, holding period
	PortfolioWeights []float64   `json:"portfolio_weights"` // 10 elements: portfolio allocation
	NewsSentiment    []float64   `json:"news_sentiment"`    // 3 elements: sentiment scores
}

// StockRecommendationResponse represents a response from stock recommendation
// FUTURE: Response will come from FastAPI ML service
type StockRecommendationResponse struct {
	RecommendationScore float64 `json:"recommendation_score"`
	Confidence          float64 `json:"confidence,omitempty"`
	Status              string  `json:"status"`
}

package ml

import "wealthscope-ai/internal/entity"

// ExtractTicker finds a stock ticker in a message.
// Supports $AAPL, plain tickers, and common company names (via entity extraction).
func ExtractTicker(message string) string {
	return entity.Extract(message).PrimaryTicker
}

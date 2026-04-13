package service

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/openai"
	"wealthscope-ai/internal/rag"
)

// ChatServiceInterface allows mocking in unit tests
type ChatServiceInterface interface {
	ProcessMessage(sessionID string, message string) (string, error)
}

type chatService struct{}

func NewChatService() ChatServiceInterface {
	return &chatService{}
}

func (s *chatService) ProcessMessage(sessionID string, message string) (string, error) {
	return ProcessMessage(sessionID, message)
}

func ProcessMessage(sessionID string, message string) (string, error) {

	ent := entity.Extract(message)
	intentResult := ml.DetectIntent(message)
	sentiment := ml.AnalyzeSentiment(message)
	enriched := message

	rctx := rag.RetrievalContextFromEntity(ent)
	chunks := rag.RetrieveWithContext(message, rctx, 3)
	if len(chunks) > 0 {
		enriched += "\n\n[Retrieved knowledge — use only if relevant]"
		for _, ch := range chunks {
			enriched += fmt.Sprintf("\n- [%s] %s", ch.Topic, strings.TrimSpace(ch.Content))
		}
	}

	if shouldEnrichMarket(intentResult.Intent, intentResult.Ticker) {
		ticker := intentResult.Ticker

		quote, err := market.GetStockQuote(ticker)
		if err == nil {
			enriched += fmt.Sprintf(
				"\n\n[Live Price Data for %s]\nPrice: $%s | High: $%s | Low: $%s | Change: %s (%s) | Volume: %s",
				quote.Symbol, quote.Price, quote.High, quote.Low,
				quote.Change, quote.ChangePercent, quote.Volume,
			)
		}

		overview, err := market.GetCompanyOverview(ticker)
		if err == nil {
			enriched += fmt.Sprintf(
				"\n\n[Company Fundamentals]\nName: %s | Sector: %s | Industry: %s\nMarket Cap: $%s | P/E: %s | EPS: %s | Beta: %s\n52W High: $%s | 52W Low: $%s | Dividend Yield: %s | Profit Margin: %s\nDescription: %s",
				overview.Name, overview.Sector, overview.Industry,
				overview.MarketCap, overview.PERatio, overview.EPS, overview.Beta,
				overview.Week52High, overview.Week52Low,
				overview.DivYield, overview.ProfitMargin,
				overview.Description,
			)
		}

		news, err := market.GetMarketNews(ticker)
		if err == nil && len(news) > 0 {
			enriched += fmt.Sprintf("\n\n[Latest News for %s]", ticker)
			for i, article := range news {
				if i >= 3 {
					break
				}
				enriched += fmt.Sprintf(
					"\n%d. %s — %s (%s)",
					i+1, article.Title, article.Source.Name, article.PublishedAt,
				)
			}
		}
	}

	enriched += fmt.Sprintf(
		"\n\n[System Context]\nIntent: %s | Ticker: %s | Sentiment: %s | Confidence: %.2f",
		intentResult.Intent,
		strings.ToUpper(intentResult.Ticker),
		sentiment,
		intentResult.Confidence,
	)

	return openai.CallOpenAI(sessionID, enriched)
}

func shouldEnrichMarket(intent ml.Intent, ticker string) bool {
	if ticker == "" {
		return false
	}
	switch intent {
	case ml.IntentStockPrice, ml.IntentRiskAnalysis, ml.IntentMarketNews:
		return true
	default:
		return false
	}
}

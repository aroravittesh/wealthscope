package service

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/chatprompt"
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

	var knowledgeLines []string
	rctx := rag.RetrievalContextFromEntity(ent)
	chunks := rag.RetrieveWithContext(message, rctx, 3)
	for _, ch := range chunks {
		knowledgeLines = append(knowledgeLines,
			fmt.Sprintf("[%s] %s", ch.Topic, strings.TrimSpace(ch.Content)))
	}

	var liveMarket strings.Builder
	var newsBody strings.Builder

	if shouldEnrichMarket(intentResult.Intent, intentResult.Ticker) {
		ticker := intentResult.Ticker

		quote, err := market.GetStockQuote(ticker)
		if err == nil {
			fmt.Fprintf(&liveMarket, "Quote (%s): price $%s | high $%s | low $%s | change %s (%s) | volume %s\n",
				quote.Symbol, quote.Price, quote.High, quote.Low,
				quote.Change, quote.ChangePercent, quote.Volume)
		}

		overview, err := market.GetCompanyOverview(ticker)
		if err == nil {
			fmt.Fprintf(&liveMarket, "Fundamentals: %s | sector %s | industry %s | market cap $%s | P/E %s | EPS %s | beta %s | 52w high $%s | 52w low $%s | div yield %s | profit margin %s\nDescription (excerpt): %s",
				overview.Name, overview.Sector, overview.Industry,
				overview.MarketCap, overview.PERatio, overview.EPS, overview.Beta,
				overview.Week52High, overview.Week52Low,
				overview.DivYield, overview.ProfitMargin,
				truncateRunes(overview.Description, 400),
			)
		}

		news, err := market.GetMarketNews(ticker)
		if err == nil && len(news) > 0 {
			for i, article := range news {
				if i >= 3 {
					break
				}
				fmt.Fprintf(&newsBody, "%d. %s — source %s (%s)\n",
					i+1, article.Title, article.Source.Name, article.PublishedAt)
			}
		}
	}

	enriched := chatprompt.BuildUserContent(chatprompt.EnvelopeInput{
		UserMessage:      message,
		KnowledgeLines:   knowledgeLines,
		LiveMarketBody:   strings.TrimSpace(liveMarket.String()),
		NewsBody:         strings.TrimSpace(newsBody.String()),
		PortfolioBody:    "",
		Intent:           string(intentResult.Intent),
		Ticker:           intentResult.Ticker,
		Sentiment:        string(sentiment),
		IntentConfidence: intentResult.Confidence,
	})

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

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

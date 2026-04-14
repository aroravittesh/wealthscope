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

// EnvelopeMarketFetchers overrides live market/news HTTP for BuildEnvelopeInputForChat (tests).
type EnvelopeMarketFetchers struct {
	GetStockQuote      func(symbol string) (*market.GlobalQuote, error)
	GetCompanyOverview func(symbol string) (*market.CompanyOverview, error)
	GetMarketNews      func(symbol string) ([]market.NewsItem, error)
}

var envelopeMarketFetchers *EnvelopeMarketFetchers

// SetEnvelopeMarketFetchersForTest installs fetcher overrides; pass nil to clear.
func SetEnvelopeMarketFetchersForTest(f *EnvelopeMarketFetchers) (cleanup func()) {
	prev := envelopeMarketFetchers
	envelopeMarketFetchers = f
	return func() { envelopeMarketFetchers = prev }
}

func resolveQuoteFn() func(string) (*market.GlobalQuote, error) {
	if envelopeMarketFetchers != nil && envelopeMarketFetchers.GetStockQuote != nil {
		return envelopeMarketFetchers.GetStockQuote
	}
	return market.GetStockQuote
}

func resolveOverviewFn() func(string) (*market.CompanyOverview, error) {
	if envelopeMarketFetchers != nil && envelopeMarketFetchers.GetCompanyOverview != nil {
		return envelopeMarketFetchers.GetCompanyOverview
	}
	return market.GetCompanyOverview
}

func resolveNewsFn() func(string) ([]market.NewsItem, error) {
	if envelopeMarketFetchers != nil && envelopeMarketFetchers.GetMarketNews != nil {
		return envelopeMarketFetchers.GetMarketNews
	}
	return market.GetMarketNews
}

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
	in := BuildEnvelopeInputForChat(message)
	enriched := chatprompt.BuildUserContent(in)
	return openai.CallOpenAI(sessionID, enriched)
}

// BuildEnvelopeInputForChat runs retrieval, optional live market enrichment, and intent/sentiment metadata.
// Exposed for tests (no OpenAI call).
func BuildEnvelopeInputForChat(message string) chatprompt.EnvelopeInput {
	ent := entity.Extract(message)
	intentResult := ml.DetectIntent(message)
	sentiment := ml.AnalyzeSentiment(message)

	var knowledgeLines []string
	var qaKnowledgeLines []string
	rctx := rag.RetrievalContextFromEntity(ent)
	chunks := rag.RetrieveWithContext(message, rctx, 3)
	for _, ch := range chunks {
		knowledgeLines = append(knowledgeLines,
			fmt.Sprintf("[%s] %s", ch.Topic, strings.TrimSpace(ch.Content)))
	}

	qaChunks := rag.RetrieveQAWithContext(message, rctx, 3)
	for _, ch := range qaChunks {
		q, a := rag.ChunkQAPair(ch)
		qaKnowledgeLines = append(qaKnowledgeLines, rag.FormatQAKnowledgeLine(ch, q, a))
	}

	var liveMarket strings.Builder
	var newsBody strings.Builder

	if shouldEnrichMarket(intentResult.Intent, intentResult.Ticker) {
		ticker := intentResult.Ticker

		quote, err := resolveQuoteFn()(ticker)
		if err == nil {
			fmt.Fprintf(&liveMarket, "Quote (%s): price $%s | high $%s | low $%s | change %s (%s) | volume %s\n",
				quote.Symbol, quote.Price, quote.High, quote.Low,
				quote.Change, quote.ChangePercent, quote.Volume)
		}

		overview, err := resolveOverviewFn()(ticker)
		if err == nil {
			fmt.Fprintf(&liveMarket, "Fundamentals: %s | sector %s | industry %s | market cap $%s | P/E %s | EPS %s | beta %s | 52w high $%s | 52w low $%s | div yield %s | profit margin %s\nDescription (excerpt): %s",
				overview.Name, overview.Sector, overview.Industry,
				overview.MarketCap, overview.PERatio, overview.EPS, overview.Beta,
				overview.Week52High, overview.Week52Low,
				overview.DivYield, overview.ProfitMargin,
				truncateRunes(overview.Description, 400),
			)
		}

		news, err := resolveNewsFn()(ticker)
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

	return chatprompt.EnvelopeInput{
		UserMessage:      message,
		KnowledgeLines:   knowledgeLines,
		QAKnowledgeLines: qaKnowledgeLines,
		LiveMarketBody:   strings.TrimSpace(liveMarket.String()),
		NewsBody:         strings.TrimSpace(newsBody.String()),
		PortfolioBody:    "",
		Intent:           string(intentResult.Intent),
		Ticker:           intentResult.Ticker,
		Sentiment:        string(sentiment),
		IntentConfidence: intentResult.Confidence,
	}
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

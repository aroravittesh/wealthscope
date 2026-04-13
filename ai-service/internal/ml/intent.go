package ml

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/intentremote"
)

type Intent string

const (
	IntentStockPrice    Intent = "STOCK_PRICE"
	IntentRiskAnalysis  Intent = "RISK_ANALYSIS"
	IntentMarketNews    Intent = "MARKET_NEWS"
	IntentPortfolioTip  Intent = "PORTFOLIO_TIP"
	IntentGeneralMarket Intent = "GENERAL_MARKET"
	IntentUnknown       Intent = "UNKNOWN"
)

type IntentResult struct {
	Intent     Intent
	Ticker     string
	Confidence float64
}

// IntentConfig controls remote classification; empty ClassifierBaseURL skips HTTP.
type IntentConfig struct {
	ClassifierBaseURL string
	Client            *http.Client
}

// DefaultIntentConfig reads INTENT_CLASSIFIER_URL (optional) and sets an HTTP client.
func DefaultIntentConfig() IntentConfig {
	u := strings.TrimSpace(os.Getenv("INTENT_CLASSIFIER_URL"))
	return IntentConfig{
		ClassifierBaseURL: strings.TrimSuffix(u, "/"),
		Client:            &http.Client{Timeout: 5 * time.Second},
	}
}

// keyword maps for fallback intent detection
var intentKeywords = map[Intent][]string{
	IntentStockPrice:    {"price", "trading at", "current price", "stock price", "how much is", "quote"},
	IntentRiskAnalysis:  {"risk", "volatile", "volatility", "safe", "dangerous", "should i buy", "analysis"},
	IntentMarketNews:    {"news", "latest", "today", "happening", "update", "recent"},
	IntentPortfolioTip:  {"portfolio", "diversify", "allocate", "holdings", "invest", "suggestion", "recommend"},
	IntentGeneralMarket: {"market", "s&p", "nasdaq", "dow", "index", "bull", "bear", "recession"},
}

// DetectIntent tries the TF-IDF service first (if configured), then keyword fallback. Ticker always comes from entity extraction.
func DetectIntent(message string) IntentResult {
	return DetectIntentWithConfig(context.Background(), DefaultIntentConfig(), message)
}

// DetectIntentWithConfig is the injectable entrypoint for tests and custom HTTP clients.
func DetectIntentWithConfig(ctx context.Context, cfg IntentConfig, message string) IntentResult {
	ent := entity.Extract(message)
	if cfg.ClassifierBaseURL != "" {
		client := cfg.Client
		if client == nil {
			client = &http.Client{Timeout: 5 * time.Second}
		}
		if remote, ok := intentremote.Classify(ctx, client, cfg.ClassifierBaseURL, message); ok {
			if intent, valid := parseIntent(remote.Intent); valid {
				return IntentResult{
					Intent:     intent,
					Confidence: remote.Confidence,
					Ticker:     ent.PrimaryTicker,
				}
			}
		}
	}
	r := keywordIntentScore(message)
	r.Ticker = ent.PrimaryTicker
	return r
}

// DetectIntentKeywords uses only the legacy keyword overlap scorer plus entity extraction.
func DetectIntentKeywords(message string) IntentResult {
	ent := entity.Extract(message)
	r := keywordIntentScore(message)
	r.Ticker = ent.PrimaryTicker
	return r
}

func parseIntent(s string) (Intent, bool) {
	switch Intent(s) {
	case IntentStockPrice, IntentRiskAnalysis, IntentMarketNews,
		IntentPortfolioTip, IntentGeneralMarket, IntentUnknown:
		return Intent(s), true
	default:
		return IntentUnknown, false
	}
}

func keywordIntentScore(message string) IntentResult {
	lower := strings.ToLower(message)
	best := IntentResult{Intent: IntentUnknown, Confidence: 0.0}

	for intent, keywords := range intentKeywords {
		score := 0.0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				score += 1.0 / float64(len(keywords))
			}
		}
		if score > best.Confidence {
			best.Intent = intent
			best.Confidence = score
		}
	}
	return best
}

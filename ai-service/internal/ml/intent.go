package ml

import (
	"context"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/explain"
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

// Re-exported source labels (canonical strings shared with internal/explain).
const (
	IntentSourceRemote          = explain.IntentSourceRemote
	IntentSourceRemoteFallback  = explain.IntentSourceRemoteFallback
	IntentSourceLowConfFallback = explain.IntentSourceLowConfFallback
	IntentSourceKeyword         = explain.IntentSourceKeyword
)

// IntentResult is the outcome of intent classification. Source, MatchedKeywords
// and Explanation are additive explainability fields and are always populated.
type IntentResult struct {
	Intent          Intent               `json:"Intent"`
	Ticker          string               `json:"Ticker"`
	Confidence      float64              `json:"Confidence"`
	Source          string               `json:"source,omitempty"`
	MatchedKeywords []string             `json:"matched_keywords,omitempty"`
	Explanation     *explain.Explanation `json:"explanation,omitempty"`
}

// IntentConfig controls remote classification; empty ClassifierBaseURL skips HTTP.
// MinConfidence (0..1) lets low-confidence remote predictions fall back to the
// keyword scorer; 0 disables the threshold.
type IntentConfig struct {
	ClassifierBaseURL string
	Client            *http.Client
	MinConfidence     float64
}

// DefaultIntentConfig reads INTENT_CLASSIFIER_URL and INTENT_MIN_CONFIDENCE.
func DefaultIntentConfig() IntentConfig {
	u := strings.TrimSpace(os.Getenv("INTENT_CLASSIFIER_URL"))
	return IntentConfig{
		ClassifierBaseURL: strings.TrimSuffix(u, "/"),
		Client:            &http.Client{Timeout: 5 * time.Second},
		MinConfidence:     parseMinConfidence(os.Getenv("INTENT_MIN_CONFIDENCE")),
	}
}

// parseMinConfidence clamps an env-provided threshold to [0,1]; invalid → 0.
func parseMinConfidence(raw string) float64 {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

var intentKeywords = map[Intent][]string{
	IntentStockPrice:    {"price", "trading at", "current price", "stock price", "how much is", "quote"},
	IntentRiskAnalysis:  {"risk", "volatile", "volatility", "safe", "dangerous", "should i buy", "analysis"},
	IntentMarketNews:    {"news", "latest", "today", "happening", "update", "recent"},
	IntentPortfolioTip:  {"portfolio", "diversify", "allocate", "holdings", "invest", "suggestion", "recommend"},
	IntentGeneralMarket: {"market", "s&p", "nasdaq", "dow", "index", "bull", "bear", "recession"},
}

// DetectIntent tries the TF-IDF service first (if configured), then keyword fallback.
func DetectIntent(message string) IntentResult {
	return DetectIntentWithConfig(context.Background(), DefaultIntentConfig(), message)
}

// DetectIntentWithConfig is the injectable entrypoint for tests and custom HTTP clients.
//
// Source taxonomy:
//   - IntentSourceRemote          → remote classifier accepted
//   - IntentSourceLowConfFallback → remote returned a valid label below MinConfidence
//   - IntentSourceRemoteFallback  → remote unreachable / invalid response
//   - IntentSourceKeyword         → no remote URL configured
func DetectIntentWithConfig(ctx context.Context, cfg IntentConfig, message string) IntentResult {
	ent := entity.Extract(message)

	if cfg.ClassifierBaseURL != "" {
		client := cfg.Client
		if client == nil {
			client = &http.Client{Timeout: 5 * time.Second}
		}
		remote, ok := intentremote.Classify(ctx, client, cfg.ClassifierBaseURL, message)
		if !ok {
			return finalizeKeyword(message, ent.PrimaryTicker, IntentSourceRemoteFallback)
		}
		intentVal, valid := parseIntent(remote.Intent)
		if !valid {
			return finalizeKeyword(message, ent.PrimaryTicker, IntentSourceRemoteFallback)
		}
		if remote.Confidence < cfg.MinConfidence {
			res := finalizeKeyword(message, ent.PrimaryTicker, IntentSourceLowConfFallback)
			// Preserve the remote confidence in the explanation copy so
			// operators can see what the classifier actually returned.
			if res.Explanation != nil {
				res.Explanation.Reasons = append(
					[]string{},
					append([]string{
						"Remote classifier returned " + remote.Intent +
							" with confidence " + formatFloat(remote.Confidence),
					}, res.Explanation.Reasons...)...,
				)
			}
			return res
		}
		result := IntentResult{
			Intent:     intentVal,
			Ticker:     ent.PrimaryTicker,
			Confidence: remote.Confidence,
			Source:     IntentSourceRemote,
		}
		exp := explain.BuildIntentExplanation(string(intentVal), ent.PrimaryTicker, IntentSourceRemote, remote.Confidence, nil)
		result.Explanation = &exp
		return result
	}

	return finalizeKeyword(message, ent.PrimaryTicker, IntentSourceKeyword)
}

// DetectIntentKeywords uses only the legacy keyword overlap scorer plus entity
// extraction. Used by tests and as a deterministic fallback.
func DetectIntentKeywords(message string) IntentResult {
	ent := entity.Extract(message)
	return finalizeKeyword(message, ent.PrimaryTicker, IntentSourceKeyword)
}

func finalizeKeyword(message, ticker, source string) IntentResult {
	scored := keywordIntentScore(message)
	res := IntentResult{
		Intent:          scored.intent,
		Ticker:          ticker,
		Confidence:      scored.confidence,
		Source:          source,
		MatchedKeywords: scored.matchedKeywords,
	}
	exp := explain.BuildIntentExplanation(string(scored.intent), ticker, source, scored.confidence, scored.matchedKeywords)
	res.Explanation = &exp
	return res
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

type keywordScoreResult struct {
	intent          Intent
	confidence      float64
	matchedKeywords []string
}

// keywordIntentScore deterministically picks the intent with the highest
// keyword-overlap score. Iteration order is fixed by sorting intents so that
// the matchedKeywords slice is reproducible across runs.
func keywordIntentScore(message string) keywordScoreResult {
	lower := strings.ToLower(message)

	intents := make([]Intent, 0, len(intentKeywords))
	for k := range intentKeywords {
		intents = append(intents, k)
	}
	sort.Slice(intents, func(i, j int) bool { return intents[i] < intents[j] })

	best := keywordScoreResult{intent: IntentUnknown, confidence: 0, matchedKeywords: nil}
	for _, intent := range intents {
		keywords := intentKeywords[intent]
		score := 0.0
		var matches []string
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				score += 1.0 / float64(len(keywords))
				matches = append(matches, kw)
			}
		}
		if score > best.confidence {
			best.intent = intent
			best.confidence = score
			best.matchedKeywords = matches
		}
	}
	return best
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

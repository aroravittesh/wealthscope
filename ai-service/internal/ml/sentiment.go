package ml

import (
	"strings"

	"wealthscope-ai/internal/finsentiment"
)

type Sentiment string

const (
	SentimentBullish Sentiment = "BULLISH"
	SentimentBearish Sentiment = "BEARISH"
	SentimentNeutral Sentiment = "NEUTRAL"
	// SentimentMixed is emitted at aggregation time when articles disagree
	// strongly enough that no single direction is meaningful.
	SentimentMixed Sentiment = "MIXED"
)

// Legacy lexicons kept for callers that still want raw keyword counts.
// The richer finance-aware scoring lives in internal/finsentiment.
var bullishWords = []string{
	"bullish", "buy", "surge", "rally", "gain", "growth",
	"outperform", "upgrade", "strong", "positive", "up",
}

var bearishWords = []string{
	"bearish", "sell", "drop", "fall", "loss", "decline",
	"underperform", "downgrade", "weak", "negative", "down",
}

// LexicalSentimentScores counts bullish vs bearish keyword hits using the
// legacy small lexicon. Preserved for backward compatibility; new code
// should use finsentiment.ScoreText / ScoreArticle for finance-aware scoring.
func LexicalSentimentScores(text string) (bullScore, bearScore int) {
	lower := strings.ToLower(text)
	for _, w := range bullishWords {
		if strings.Contains(lower, w) {
			bullScore++
		}
	}
	for _, w := range bearishWords {
		if strings.Contains(lower, w) {
			bearScore++
		}
	}
	return bullScore, bearScore
}

// AnalyzeSentiment classifies an arbitrary text using the finance-aware scorer.
// Returns BULLISH / BEARISH / NEUTRAL only — MIXED is reserved for aggregation.
func AnalyzeSentiment(text string) Sentiment {
	if strings.TrimSpace(text) == "" {
		return SentimentNeutral
	}
	score := finsentiment.ScoreText(text)
	switch score.Bucket() {
	case finsentiment.Bullish:
		return SentimentBullish
	case finsentiment.Bearish:
		return SentimentBearish
	default:
		return SentimentNeutral
	}
}

package ml

import "strings"

type Sentiment string

const (
	SentimentBullish Sentiment = "BULLISH"
	SentimentBearish Sentiment = "BEARISH"
	SentimentNeutral Sentiment = "NEUTRAL"
)

var bullishWords = []string{
	"bullish", "buy", "surge", "rally", "gain", "growth",
	"outperform", "upgrade", "strong", "positive", "up",
}

var bearishWords = []string{
	"bearish", "sell", "drop", "fall", "loss", "decline",
	"underperform", "downgrade", "weak", "negative", "down",
}

// LexicalSentimentScores counts bullish vs bearish keyword hits (baseline lexicon).
// Used by AnalyzeSentiment and news aggregation; easy to swap for a classifier later.
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

func AnalyzeSentiment(text string) Sentiment {
	bullScore, bearScore := LexicalSentimentScores(text)
	if bullScore > bearScore {
		return SentimentBullish
	} else if bearScore > bullScore {
		return SentimentBearish
	}
	return SentimentNeutral
}

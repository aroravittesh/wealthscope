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

func AnalyzeSentiment(text string) Sentiment {
    lower := strings.ToLower(text)
    bullScore, bearScore := 0, 0

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

    if bullScore > bearScore {
        return SentimentBullish
    } else if bearScore > bullScore {
        return SentimentBearish
    }
    return SentimentNeutral
}
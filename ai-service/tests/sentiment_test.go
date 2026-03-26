package tests

import (
    "testing"
    "wealthscope-ai/internal/ml"
)

func TestSentiment_Bullish(t *testing.T) {
    result := ml.AnalyzeSentiment("AAPL is surging with strong growth")
    if result != ml.SentimentBullish {
        t.Fatalf("expected BULLISH got %s", result)
    }
}

func TestSentiment_Bearish(t *testing.T) {
    result := ml.AnalyzeSentiment("TSLA is dropping with weak earnings")
    if result != ml.SentimentBearish {
        t.Fatalf("expected BEARISH got %s", result)
    }
}

func TestSentiment_Neutral(t *testing.T) {
    result := ml.AnalyzeSentiment("The company released its quarterly report")
    if result != ml.SentimentNeutral {
        t.Fatalf("expected NEUTRAL got %s", result)
    }
}

func TestSentiment_BullishOverBearish(t *testing.T) {
    result := ml.AnalyzeSentiment("Strong rally and surge in gains")
    if result != ml.SentimentBullish {
        t.Fatalf("expected BULLISH got %s", result)
    }
}

func TestSentiment_EmptyText(t *testing.T) {
    result := ml.AnalyzeSentiment("")
    if result != ml.SentimentNeutral {
        t.Fatalf("expected NEUTRAL for empty text got %s", result)
    }
}
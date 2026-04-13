package tests

import (
	"strings"
	"testing"

	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/newsentiment"
)

func TestNewsSentiment_AggregateBullish(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Apple stock rally continues with strong gains", Description: "Surge in price after upgrade"},
		{Title: "AAPL outperforms with positive outlook", Description: "Growth story intact"},
	}
	r := newsentiment.Aggregate("AAPL", articles)
	if r.OverallSentiment != string(ml.SentimentBullish) {
		t.Fatalf("want BULLISH got %s", r.OverallSentiment)
	}
	if r.ArticleCount != 2 {
		t.Fatalf("article count %d", r.ArticleCount)
	}
	if r.Confidence <= 0 || r.Confidence > 1 {
		t.Fatalf("confidence %f", r.Confidence)
	}
	if r.TopPositiveArticle == "" {
		t.Fatal("expected top positive title")
	}
	if r.Summary == "" || !strings.Contains(r.Summary, "not investment advice") {
		t.Fatalf("summary missing disclaimer: %q", r.Summary)
	}
}

func TestNewsSentiment_AggregateBearish(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Shares drop on weak earnings", Description: "Stock falls after downgrade"},
		{Title: "Bearish outlook as losses mount", Description: "Decline continues"},
	}
	r := newsentiment.Aggregate("TSLA", articles)
	if r.OverallSentiment != string(ml.SentimentBearish) {
		t.Fatalf("want BEARISH got %s", r.OverallSentiment)
	}
	if r.TopNegativeArticle == "" {
		t.Fatal("expected top negative title")
	}
}

func TestNewsSentiment_AggregateNeutral(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Company schedules quarterly report", Description: "The firm will release figures next week"},
		{Title: "Analyst meeting announced", Description: "Investors await commentary"},
	}
	r := newsentiment.Aggregate("MSFT", articles)
	if r.OverallSentiment != string(ml.SentimentNeutral) {
		t.Fatalf("want NEUTRAL got %s", r.OverallSentiment)
	}
}

func TestNewsSentiment_EmptyArticles(t *testing.T) {
	r := newsentiment.Aggregate("NVDA", nil)
	if r.OverallSentiment != string(ml.SentimentNeutral) {
		t.Fatalf("want NEUTRAL got %s", r.OverallSentiment)
	}
	if r.ArticleCount != 0 {
		t.Fatalf("want 0 articles got %d", r.ArticleCount)
	}
	if r.Confidence != 0 {
		t.Fatalf("want 0 confidence got %f", r.Confidence)
	}
	if r.TopPositiveArticle != "" || r.TopNegativeArticle != "" {
		t.Fatal("expected empty top articles")
	}
	if !strings.Contains(r.Summary, "No recent articles") {
		t.Fatalf("summary: %q", r.Summary)
	}
}

func TestLexicalSentimentScores_Exported(t *testing.T) {
	b, be := ml.LexicalSentimentScores("rally surge gain")
	if b <= be {
		t.Fatalf("expected more bull than bear hits: %d %d", b, be)
	}
}

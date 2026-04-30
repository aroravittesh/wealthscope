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

func TestNewsSentiment_AggregateMixedBullAndBearPopulatesBothTops(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Rally surge gain upside momentum continues", Description: "Strong gains and growth outlook"},
		{Title: "Crash plunge collapse downside losses deepen", Description: "Bearish slide and steep decline"},
	}
	r := newsentiment.Aggregate("XOM", articles)
	if r.ArticleCount != 2 {
		t.Fatalf("article count %d", r.ArticleCount)
	}
	if r.TopPositiveArticle == "" || r.TopNegativeArticle == "" {
		t.Fatalf("expected both top article fields set, got pos=%q neg=%q", r.TopPositiveArticle, r.TopNegativeArticle)
	}
	switch r.OverallSentiment {
	case string(ml.SentimentBullish), string(ml.SentimentBearish), string(ml.SentimentNeutral), string(ml.SentimentMixed):
	default:
		t.Fatalf("unexpected overall %q", r.OverallSentiment)
	}
}

// MIXED is emitted when articles disagree strongly enough that neither side wins.
func TestNewsSentiment_AggregateMixedClassification(t *testing.T) {
	articles := []market.NewsItem{
		{
			Title:       "Apple beat estimates and raised guidance for the year",
			Description: "Record high revenue with strong dividend increase announced",
		},
		{
			Title:       "Apple missed estimates and cut guidance amid product recall",
			Description: "Profit warning and earnings miss send shares to a record low",
		},
	}
	r := newsentiment.Aggregate("AAPL", articles)
	if r.OverallSentiment != string(ml.SentimentMixed) {
		t.Fatalf("want MIXED for split articles, got %s (summary: %s)", r.OverallSentiment, r.Summary)
	}
	if r.Confidence > 0.55 {
		t.Fatalf("MIXED confidence must be capped at 0.55, got %f", r.Confidence)
	}
	if r.TopPositiveArticle == "" || r.TopNegativeArticle == "" {
		t.Fatalf("MIXED must surface both extremes, got pos=%q neg=%q", r.TopPositiveArticle, r.TopNegativeArticle)
	}
}

// Finance-specific phrases must be honored over generic word counts.
func TestNewsSentiment_FinancePhrasesDriveBullish(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Apple beat estimates", Description: "Raised guidance after blowout quarter"},
	}
	r := newsentiment.Aggregate("AAPL", articles)
	if r.OverallSentiment != string(ml.SentimentBullish) {
		t.Fatalf("want BULLISH from finance phrases, got %s", r.OverallSentiment)
	}
	if r.Confidence <= 0 {
		t.Fatalf("expected confidence > 0, got %f", r.Confidence)
	}
}

// Negation should not let an article be misclassified as bearish.
func TestNewsSentiment_NegationDoesNotInvertSignal(t *testing.T) {
	articles := []market.NewsItem{
		{Title: "Analysts say no slowdown ahead, growth remains strong", Description: "Outlook robust"},
	}
	r := newsentiment.Aggregate("MSFT", articles)
	if r.OverallSentiment == string(ml.SentimentBearish) {
		t.Fatalf("negation should prevent BEARISH classification, got %s (summary: %s)", r.OverallSentiment, r.Summary)
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

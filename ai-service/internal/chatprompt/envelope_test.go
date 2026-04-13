package chatprompt

import (
	"strings"
	"testing"
)

func TestBuildUserContent_SectionOrderAndHeaders(t *testing.T) {
	out := BuildUserContent(EnvelopeInput{
		UserMessage:      "What is beta?",
		KnowledgeLines:   []string{"[risk] Beta measures volatility."},
		LiveMarketBody:   "Quote: AAPL $180",
		NewsBody:         "1. Headline",
		PortfolioBody:    "",
		Intent:           "GENERAL_MARKET",
		Ticker:           "",
		Sentiment:        "NEUTRAL",
		IntentConfidence: 0.5,
	})

	ik := strings.Index(out, SectionKnowledge)
	il := strings.Index(out, SectionLiveMarket)
	in := strings.Index(out, SectionNews)
	ip := strings.Index(out, SectionPortfolio)
	is := strings.Index(out, SectionSystem)
	if ik < 0 || il < 0 || in < 0 || ip < 0 || is < 0 {
		t.Fatalf("missing section: k=%d l=%d n=%d p=%d s=%d", ik, il, in, ip, is)
	}
	if !(ik < il && il < in && in < ip && ip < is) {
		t.Fatal("sections should appear in order: knowledge → live market → news → portfolio → system")
	}
	if !strings.Contains(out, "What is beta?") {
		t.Fatal("original user message missing")
	}
}

func TestBuildUserContent_MissingKnowledge(t *testing.T) {
	out := BuildUserContent(EnvelopeInput{
		UserMessage: "Hello",
		Intent:      "UNKNOWN",
		Sentiment:   "NEUTRAL",
	})
	if !strings.Contains(out, "No curated knowledge snippets") {
		t.Fatalf("expected empty knowledge note: %s", out)
	}
}

func TestBuildUserContent_MissingMarketAndNews(t *testing.T) {
	out := BuildUserContent(EnvelopeInput{
		UserMessage: "Explain P/E",
		Intent:      "UNKNOWN",
		Sentiment:   "NEUTRAL",
	})
	if !strings.Contains(out, "No live market data was attached") {
		t.Fatal("expected missing market note")
	}
	if !strings.Contains(out, "No news headlines were attached") {
		t.Fatal("expected missing news note")
	}
	if !strings.Contains(out, "No portfolio holdings") {
		t.Fatal("expected default portfolio note")
	}
}

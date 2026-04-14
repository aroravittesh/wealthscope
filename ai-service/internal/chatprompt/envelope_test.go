package chatprompt

import (
	"strings"
	"testing"
)

func TestBuildUserContent_SectionOrderAndHeaders(t *testing.T) {
	out := BuildUserContent(EnvelopeInput{
		UserMessage:      "What is beta?",
		KnowledgeLines:   []string{"[risk] Beta measures volatility."},
		QAKnowledgeLines: []string{"[QA0001 | Stock Basics / x] Q: What is beta? | A: Beta measures market sensitivity."},
		LiveMarketBody:   "Quote: AAPL $180",
		NewsBody:         "1. Headline",
		PortfolioBody:    "",
		Intent:           "GENERAL_MARKET",
		Ticker:           "",
		Sentiment:        "NEUTRAL",
		IntentConfidence: 0.5,
	})

	ik := strings.Index(out, SectionKnowledge)
	iq := strings.Index(out, SectionQAKnowledge)
	il := strings.Index(out, SectionLiveMarket)
	in := strings.Index(out, SectionNews)
	ip := strings.Index(out, SectionPortfolio)
	is := strings.Index(out, SectionSystem)
	if ik < 0 || iq < 0 || il < 0 || in < 0 || ip < 0 || is < 0 {
		t.Fatalf("missing section: k=%d q=%d l=%d n=%d p=%d s=%d", ik, iq, il, in, ip, is)
	}
	if !(ik < iq && iq < il && il < in && in < ip && ip < is) {
		t.Fatal("sections should appear in order: knowledge → QA knowledge → live market → news → portfolio → system")
	}
	if !strings.Contains(out, "QA0001") {
		t.Fatal("expected QA line in envelope")
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
	if !strings.Contains(out, "No matching Q&A knowledge rows") {
		t.Fatal("expected empty QA knowledge note")
	}
}

func TestBuildUserContent_SectionHeadersAppearOnce(t *testing.T) {
	out := BuildUserContent(EnvelopeInput{
		UserMessage:      "Hello",
		KnowledgeLines:   []string{"[x] y"},
		QAKnowledgeLines: []string{"[QA1 | A / B] Q: q | A: a"},
		LiveMarketBody:   "m",
		NewsBody:         "n",
		PortfolioBody:    "p",
		Intent:           "X",
		Sentiment:        "NEUTRAL",
	})
	for _, title := range []string{
		SectionKnowledge, SectionQAKnowledge, SectionLiveMarket,
		SectionNews, SectionPortfolio, SectionSystem,
	} {
		if strings.Count(out, title) != 1 {
			t.Fatalf("expected exactly one %q, count=%d", title, strings.Count(out, title))
		}
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

package websearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wealthscope-ai/internal/entity"
)

// ---- Decide ----

func TestDecide_TimeSensitiveQueryWithTickerTriggers(t *testing.T) {
	ent := entity.EntityResult{PrimaryTicker: "TSLA"}
	d := Decide("What is the latest news on Tesla today?", "MARKET_NEWS", ent)
	if !d.Use {
		t.Fatalf("expected Use=true, reason=%s", d.Reason)
	}
	if !strings.Contains(d.Query, "TSLA") {
		t.Fatalf("query should mention ticker: %q", d.Query)
	}
	if !strings.Contains(d.Reason, "TSLA") {
		t.Fatalf("reason should reference ticker: %q", d.Reason)
	}
}

func TestDecide_EvergreenQuestionDoesNotTrigger(t *testing.T) {
	cases := []string{
		"What is beta?",
		"Define diversification",
		"Explain how P/E ratio works",
		"How does dollar cost averaging work",
	}
	for _, msg := range cases {
		t.Run(msg, func(t *testing.T) {
			d := Decide(msg, "GENERAL_MARKET", entity.EntityResult{})
			if d.Use {
				t.Fatalf("evergreen %q should not trigger; reason=%s query=%s", msg, d.Reason, d.Query)
			}
		})
	}
}

func TestDecide_ExplicitNewsPhraseTriggersEvenWithoutTimeToken(t *testing.T) {
	d := Decide("any news on Apple", "MARKET_NEWS", entity.EntityResult{PrimaryTicker: "AAPL"})
	if !d.Use {
		t.Fatalf("explicit news phrase should trigger: %+v", d)
	}
}

func TestDecide_PortfolioIntentDoesNotTrigger(t *testing.T) {
	d := Decide("How should I diversify my portfolio today", "PORTFOLIO_TIP", entity.EntityResult{})
	if d.Use {
		t.Fatalf("portfolio intent should be skipped, got %+v", d)
	}
}

func TestDecide_EmptyMessageDoesNotTrigger(t *testing.T) {
	if d := Decide("   ", "MARKET_NEWS", entity.EntityResult{}); d.Use {
		t.Fatal("empty input should not trigger")
	}
}

func TestDecide_TimeTokenWithoutTickerStillTriggers(t *testing.T) {
	d := Decide("What happened to the market today?", "GENERAL_MARKET", entity.EntityResult{})
	if !d.Use {
		t.Fatalf("expected Use=true, got %+v", d)
	}
	if d.Query == "" {
		t.Fatal("query must be non-empty")
	}
}

// ---- CleanAndRank ----

func TestCleanAndRank_DropsEmptyAndShortAndDeduplicates(t *testing.T) {
	in := []Result{
		{Title: "", Snippet: "real body of news goes here at length"},
		{Title: "Empty snippet", Snippet: ""},
		{Title: "Too short", Snippet: "tiny."},
		{Title: "Tesla beats Q1 expectations", Snippet: "Tesla reported quarterly results with revenue and margin metrics within guidance.", URL: "https://reuters.com/articles/tsla-q1", Source: "reuters.com", Score: 0.5},
		{Title: "Tesla beats Q1 expectations", Snippet: "Duplicate URL different snippet but enough characters to pass min length filter.", URL: "https://reuters.com/articles/tsla-q1", Source: "reuters.com", Score: 0.4},
		{Title: "Apple announces buyback", Snippet: "Apple announced a major share repurchase plan this quarter, lifting after-hours sentiment.", URL: "https://bloomberg.com/news/aapl-buyback", Source: "bloomberg.com", Score: 0.6},
		{Title: "Random spam page", Snippet: "Click here right now to learn the secret of investing this is a long enough snippet.", URL: "https://example.com/spam", Source: "example.com", Score: 0.95},
	}
	out := CleanAndRank(in, 3)

	if len(out) != 3 {
		t.Fatalf("want 3 cleaned results got %d", len(out))
	}
	// Trusted Bloomberg should beat the high-score spam due to boost.
	if out[0].Source != "bloomberg.com" {
		t.Fatalf("want bloomberg first, got %s", out[0].Source)
	}
	// Reuters trusted should beat spam too.
	if out[1].Source != "reuters.com" {
		t.Fatalf("want reuters second, got %s", out[1].Source)
	}
}

func TestCleanAndRank_NilInputReturnsNil(t *testing.T) {
	if got := CleanAndRank(nil, 3); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestCleanAndRank_TruncatesLongSnippets(t *testing.T) {
	long := strings.Repeat("x", maxSnippetChars+50)
	out := CleanAndRank([]Result{{Title: "Long", Snippet: long, URL: "https://reuters.com/x", Source: "reuters.com"}}, 1)
	if len(out) != 1 {
		t.Fatalf("want 1 got %d", len(out))
	}
	if !strings.HasSuffix(out[0].Snippet, "…") {
		t.Fatalf("expected ellipsis suffix on truncated snippet: %q", out[0].Snippet)
	}
}

// ---- FormatForPrompt ----

func TestFormatForPrompt_BuildsNumberedLines(t *testing.T) {
	body := FormatForPrompt([]Result{
		{Title: "A1", Snippet: "Body of article one with enough length.", Source: "reuters.com", PublishedAt: "2026-04-29"},
		{Title: "A2", Snippet: "Body of article two also long enough.", Source: "bloomberg.com"},
	})
	if !strings.HasPrefix(body, "1. A1 — reuters.com (2026-04-29):") {
		t.Fatalf("first line malformed: %q", body)
	}
	if !strings.Contains(body, "\n2. A2 — bloomberg.com:") {
		t.Fatalf("second line malformed: %q", body)
	}
}

func TestFormatForPrompt_EmptyReturnsEmpty(t *testing.T) {
	if got := FormatForPrompt(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// ---- TavilyProvider ----

func TestTavilyProvider_SuccessParsesResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected json content-type got %s", ct)
		}
		_, _ = w.Write([]byte(`{
            "results": [
                {"title":"Apple Q1","url":"https://reuters.com/x","content":"Apple reported strong revenue this quarter.","score":0.91,"published_date":"2026-04-29"},
                {"title":"AAPL price","url":"https://bloomberg.com/y","content":"AAPL up 3% intraday on news.","score":0.72,"published_date":"2026-04-28"}
            ]
        }`))
	}))
	t.Cleanup(ts.Close)

	p := NewTavilyProviderWithURL("k", ts.URL, ts.Client())
	got, err := p.Search(context.Background(), "AAPL stock latest news", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 hits got %d", len(got))
	}
	if got[0].Source != "reuters.com" {
		t.Fatalf("source not derived from URL: %s", got[0].Source)
	}
	if got[0].Score == 0 {
		t.Fatal("score not parsed")
	}
}

func TestTavilyProvider_NonOKReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`bad key`))
	}))
	t.Cleanup(ts.Close)
	p := NewTavilyProviderWithURL("k", ts.URL, ts.Client())
	if _, err := p.Search(context.Background(), "q", 3); err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestTavilyProvider_MissingAPIKeyErrors(t *testing.T) {
	p := NewTavilyProvider("", nil)
	if _, err := p.Search(context.Background(), "q", 3); err == nil {
		t.Fatal("expected missing-key error")
	}
}

// ---- Provider selection ----

func TestProviderFromEnv_StubByDefault(t *testing.T) {
	p := providerFromConfig(ProviderConfig{})
	if p.Name() != "stub" {
		t.Fatalf("want stub got %s", p.Name())
	}
}

func TestProviderFromEnv_ExplicitOff(t *testing.T) {
	if p := providerFromConfig(ProviderConfig{Provider: "off", TavilyKey: "k"}); p.Name() != "stub" {
		t.Fatalf("want stub got %s", p.Name())
	}
}

func TestProviderFromEnv_AutoTavilyWhenKeyPresent(t *testing.T) {
	if p := providerFromConfig(ProviderConfig{TavilyKey: "tv-key"}); p.Name() != "tavily" {
		t.Fatalf("want tavily got %s", p.Name())
	}
}

// ---- StubProvider ----

func TestStubProvider_AlwaysEmpty(t *testing.T) {
	got, err := StubProvider{}.Search(context.Background(), "anything", 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

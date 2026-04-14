package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wealthscope-ai/internal/chatprompt"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/rag"
)

func TestBuildEnvelopeInputForChat_IncludesKnowledgeAndQA(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	header := strings.Join([]string{
		"id", "category", "sub_category", "question", "answer", "keywords",
		"ticker", "difficulty", "source_type", "priority", "last_updated",
	}, ",")
	csv := strings.Join([]string{
		header,
		`QA9001,Diversification,corr,"What is diversification?","Spreading investments reduces concentration risk.","diversification,risk",,beginner,educational,high,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	rag.SetQADatasetPathForTest(p)
	t.Cleanup(rag.ClearQADatasetPathOverride)

	in := BuildEnvelopeInputForChat("Explain diversification and portfolio risk in simple terms")
	if len(in.KnowledgeLines) == 0 {
		t.Fatal("expected legacy financial knowledge lines from static KB")
	}
	if len(in.QAKnowledgeLines) == 0 {
		t.Fatal("expected QA CSV retrieval lines")
	}
	var sawQA bool
	for _, line := range in.QAKnowledgeLines {
		if strings.Contains(line, "QA9001") && strings.Contains(line, "diversification") {
			sawQA = true
			break
		}
	}
	if !sawQA {
		t.Fatalf("QA lines: %#v", in.QAKnowledgeLines)
	}

	full := chatprompt.BuildUserContent(in)
	if !strings.Contains(full, chatprompt.SectionQAKnowledge) {
		t.Fatal("assembled user content should include QA section when built via chatprompt")
	}
	if !strings.Contains(full, "QA9001") {
		t.Fatal("assembled prompt should include QA id")
	}
}

func TestBuildEnvelopeInputForChat_StockPriceIntentUsesStubbedMarket(t *testing.T) {
	cleanup := SetEnvelopeMarketFetchersForTest(&EnvelopeMarketFetchers{
		GetStockQuote: func(symbol string) (*market.GlobalQuote, error) {
			if symbol != "AAPL" {
				t.Fatalf("unexpected symbol %q", symbol)
			}
			return &market.GlobalQuote{
				Symbol: "AAPL", Price: "100.00", High: "101", Low: "99",
				Change: "1", ChangePercent: "1%", Volume: "1M",
			}, nil
		},
		GetCompanyOverview: func(string) (*market.CompanyOverview, error) {
			return nil, errors.New("overview skipped in test")
		},
		GetMarketNews: func(string) ([]market.NewsItem, error) {
			return nil, nil
		},
	})
	t.Cleanup(cleanup)

	in := BuildEnvelopeInputForChat("What is AAPL current stock price quote for education only")
	if !strings.Contains(in.LiveMarketBody, "Quote (AAPL)") || !strings.Contains(in.LiveMarketBody, "$100.00") {
		t.Fatalf("expected stubbed quote in live market body, got %q", in.LiveMarketBody)
	}
}

func TestBuildEnvelopeInputForChat_NoMarketEnrichmentForGenericQuestion(t *testing.T) {
	rag.SetQADatasetPathForTest(filepath.Join(t.TempDir(), "missing.csv"))
	t.Cleanup(rag.ClearQADatasetPathOverride)
	// Missing file → QA empty; still should not attach live quote without ticker intent
	in := BuildEnvelopeInputForChat("Define the concept of a limit order for education only")
	if strings.TrimSpace(in.LiveMarketBody) != "" {
		t.Fatalf("unexpected live market body: %q", in.LiveMarketBody)
	}
	if strings.TrimSpace(in.NewsBody) != "" {
		t.Fatalf("unexpected news body: %q", in.NewsBody)
	}
}

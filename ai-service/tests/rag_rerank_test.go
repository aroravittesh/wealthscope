package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/rag"
)

// --- Semantic-only finance concept query: should still find risk_analysis. ---

func TestRerank_SemanticFinanceConceptQuery(t *testing.T) {
	chunks := rag.RetrieveWithContext("how risky is the market and what is beta", rag.RetrievalContext{}, 3)
	if len(chunks) == 0 {
		t.Fatal("expected at least one knowledge chunk")
	}
	if !containsTopic(chunks, "risk_analysis") {
		t.Fatalf("expected risk_analysis in top-K, got %v", topicsOf(chunks))
	}
}

// --- Entity / ticker-heavy query: ticker presence in tags should boost the
// matching QA row above an unrelated but lexically similar one. ---

func TestRerank_EntityBoostLiftsTickerTaggedChunk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	csv := strings.Join([]string{
		strings.Join([]string{
			"id", "category", "sub_category", "question", "answer", "keywords",
			"ticker", "difficulty", "source_type", "priority", "last_updated",
		}, ","),
		// AAPL-tagged row but with weak lexical signal vs. the query
		`QA9001,Earnings,reports,"What drives quarterly earnings outlook?","Earnings outlook depends on revenue growth and margins.","earnings,outlook",AAPL,beginner,educational,medium,2026-04-12`,
		// Highly lexical row but no ticker tag
		`QA9002,Earnings,reports,"Explain quarterly earnings outlook reports","Quarterly earnings outlook reports include revenue and margins.","earnings,outlook,reports",,beginner,educational,medium,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	rag.SetQADatasetPathForTest(p)
	t.Cleanup(rag.ClearQADatasetPathOverride)

	ctx := rag.RetrievalContextFromEntity(entity.EntityResult{
		PrimaryTicker:  "AAPL",
		CompanyMatches: []string{"Apple"},
	})
	hits := rag.RetrieveQAWithContext("quarterly earnings outlook", ctx, 2)
	if len(hits) < 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].ID != "QA9001" {
		t.Fatalf("expected ticker-tagged QA9001 first, got %s (order=%v)", hits[0].ID, idsOf(hits))
	}
}

// --- Metadata priority should break ties between otherwise-equal rows. ---

func TestRerank_MetadataPriorityOrdersHigherFirst(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	csv := strings.Join([]string{
		strings.Join([]string{
			"id", "category", "sub_category", "question", "answer", "keywords",
			"ticker", "difficulty", "source_type", "priority", "last_updated",
		}, ","),
		`QA8001,Diversification,corr,"What is portfolio diversification?","Diversification spreads investments across assets.","diversification,portfolio",,beginner,educational,low,2026-04-12`,
		`QA8002,Diversification,corr,"What is portfolio diversification?","Diversification spreads investments across assets.","diversification,portfolio",,beginner,educational,high,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	rag.SetQADatasetPathForTest(p)
	t.Cleanup(rag.ClearQADatasetPathOverride)

	hits := rag.RetrieveQAWithContext("portfolio diversification", rag.RetrievalContext{}, 2)
	if len(hits) < 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].ID != "QA8002" {
		t.Fatalf("expected high-priority QA8002 first, got %s", hits[0].ID)
	}
}

// --- No-match query stays empty (eligibility filter). ---

func TestRerank_NoMatchReturnsEmpty(t *testing.T) {
	chunks := rag.RetrieveWithContext("xyzzy abracadabra grommet", rag.RetrievalContext{}, 3)
	if len(chunks) != 0 {
		t.Fatalf("expected empty results, got %v", topicsOf(chunks))
	}
}

// --- Backward compatibility: legacy Retrieve still returns a relevant doc. ---

func TestRerank_BackwardCompatLegacyRetrieve(t *testing.T) {
	docs := rag.Retrieve("dividend yield income investors", 3)
	if len(docs) == 0 {
		t.Fatal("expected docs from legacy Retrieve")
	}
	found := false
	for _, d := range docs {
		if d.Topic == "dividends" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected dividends topic, got %v", legacyTopicsOf(docs))
	}
}

// --- Top-K relevance ordering: higher final-score wins on a clean query. ---

func TestRerank_TopKRelevanceOrdering(t *testing.T) {
	chunks := rag.RetrieveWithContext("what is beta and volatility risk", rag.RetrievalContext{}, 3)
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].Topic != "risk_analysis" {
		t.Fatalf("expected risk_analysis as #1 final-ranked, got %v", topicsOf(chunks))
	}
}

// --- Multi-source: KB and QA both queried; entity boost on QA lifts the
// ticker-tagged QA row over the plain KB chunk for a ticker query. ---

func TestRerank_MultiSourceEntityCase(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "qa.csv")
	csv := strings.Join([]string{
		strings.Join([]string{
			"id", "category", "sub_category", "question", "answer", "keywords",
			"ticker", "difficulty", "source_type", "priority", "last_updated",
		}, ","),
		`QA7001,Earnings,reports,"How are dividends paid by large tech firms?","Many large tech firms pay dividends quarterly to shareholders.","dividends,tech",AAPL,beginner,educational,high,2026-04-12`,
	}, "\n")
	if err := os.WriteFile(p, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	rag.SetQADatasetPathForTest(p)
	t.Cleanup(rag.ClearQADatasetPathOverride)

	ctx := rag.RetrievalContextFromEntity(entity.EntityResult{
		PrimaryTicker:  "AAPL",
		CompanyMatches: []string{"Apple"},
	})

	kb := rag.RetrieveWithContext("dividend payments to shareholders", ctx, 3)
	qa := rag.RetrieveQAWithContext("dividend payments to shareholders", ctx, 3)

	if len(kb) == 0 {
		t.Fatal("expected KB hits")
	}
	if len(qa) == 0 {
		t.Fatal("expected QA hits")
	}
	if qa[0].ID != "QA7001" {
		t.Fatalf("expected ticker-tagged QA7001 first in QA results, got %s", qa[0].ID)
	}
	if !containsTopic(kb, "dividends") {
		t.Fatalf("expected KB dividends topic, got %v", topicsOf(kb))
	}
}

func containsTopic(chunks []rag.KnowledgeChunk, topic string) bool {
	for _, c := range chunks {
		if c.Topic == topic {
			return true
		}
	}
	return false
}

func idsOf(chunks []rag.KnowledgeChunk) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.ID
	}
	return out
}

func legacyTopicsOf(docs []rag.FinancialDocument) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.Topic
	}
	return out
}

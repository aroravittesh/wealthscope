package tests

import (
	"testing"

	"wealthscope-ai/internal/entity"
	"wealthscope-ai/internal/rag"
)

func TestRetrieveWithContext_SemanticBetaQuery(t *testing.T) {
	ctx := rag.RetrievalContext{}
	chunks := rag.RetrieveWithContext("market volatility and beta measure", ctx, 3)
	if len(chunks) == 0 {
		t.Fatal("expected semantic or lexical hits")
	}
	found := false
	for _, ch := range chunks {
		if ch.Topic == "risk_analysis" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected risk_analysis in top results, got topics %#v", topicsOf(chunks))
	}
}

func TestRetrieveWithContext_EntityBoostTicker(t *testing.T) {
	ent := entity.EntityResult{
		PrimaryTicker: "AAPL",
	}
	ctx := rag.RetrievalContextFromEntity(ent)
	chunks := rag.RetrieveWithContext("earnings and shareholder distributions", ctx, 3)
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
}

func TestRetrieve_BackwardCompatible(t *testing.T) {
	docs := rag.Retrieve("dividend yield income", 3)
	if len(docs) == 0 {
		t.Fatal("expected docs")
	}
	found := false
	for _, d := range docs {
		if d.Topic == "dividends" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected dividends topic")
	}
}

func topicsOf(chunks []rag.KnowledgeChunk) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.Topic
	}
	return out
}

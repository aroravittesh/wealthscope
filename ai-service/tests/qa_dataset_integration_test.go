package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wealthscope-ai/internal/chatprompt"
	"wealthscope-ai/internal/rag"
)

func TestRetrieveQA_RealDatasetIfPresent(t *testing.T) {
	candidates := []string{
		filepath.Join("..", "data", "qa_dataset.csv"),
		filepath.Join("..", "..", "data", "qa_dataset.csv"),
	}
	var real string
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			real = c
			break
		}
	}
	if real == "" {
		t.Skip("no repo data/qa_dataset.csv found from tests/")
	}

	rag.SetQADatasetPathForTest(real)
	t.Cleanup(rag.ClearQADatasetPathOverride)

	hits := rag.RetrieveQAWithContext("what is diversification for investors", rag.RetrievalContext{}, 5)
	if len(hits) == 0 {
		t.Fatal("expected QA hits from real dataset")
	}
	found := false
	for _, h := range hits {
		if strings.Contains(strings.ToLower(h.Topic), "diversification") {
			found = true
			break
		}
	}
	if !found {
		topics := make([]string, len(hits))
		for i := range hits {
			topics[i] = hits[i].Topic
		}
		t.Logf("top hit topics: %v", topics)
		t.Fatal("expected a diversification-related QA row in top 5")
	}
}

func TestBuildUserContent_QASectionWithRetrievalStyleLine(t *testing.T) {
	line := rag.FormatQAKnowledgeLine(
		rag.KnowledgeChunk{ID: "QA0401", Topic: "ETFs and Mutual Funds / index funds"},
		"What is benchmark drift?",
		"Benchmark drift happens when holdings diverge from the index.",
	)
	out := chatprompt.BuildUserContent(chatprompt.EnvelopeInput{
		UserMessage:      "Tell me about ETFs",
		QAKnowledgeLines: []string{line},
		Intent:           "GENERAL_MARKET",
		Sentiment:        "NEUTRAL",
	})
	if !strings.Contains(out, chatprompt.SectionQAKnowledge) {
		t.Fatal("missing QA section")
	}
	if !strings.Contains(out, "QA0401") || !strings.Contains(out, "benchmark drift") {
		snippet := out
		if len(snippet) > 400 {
			snippet = snippet[:400]
		}
		t.Fatalf("missing QA content: %s", snippet)
	}
}

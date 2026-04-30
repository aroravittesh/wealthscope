package tests

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"wealthscope-ai/internal/chatprompt"
	"wealthscope-ai/internal/rag"
	"wealthscope-ai/internal/service"
	"wealthscope-ai/internal/websearch"
)

// failingProvider always errors. Used to assert silent fallback.
type failingProvider struct{}

func (failingProvider) Name() string { return "failing" }
func (failingProvider) Search(_ context.Context, _ string, _ int) ([]websearch.Result, error) {
	return nil, errors.New("simulated provider outage")
}

// emptyQA points the QA loader at an empty directory so QA retrieval returns
// nothing — keeps the integration tests focused on the web layer.
func isolateQA(t *testing.T) {
	t.Helper()
	rag.SetQADatasetPathForTest(filepath.Join(t.TempDir(), "missing.csv"))
	t.Cleanup(rag.ClearQADatasetPathOverride)
}

func TestChatEnvelope_TimeSensitiveQueryAttachesWebContext(t *testing.T) {
	isolateQA(t)

	mock := &websearch.MockProvider{
		NameStr: "mock",
		Results: []websearch.Result{
			{
				Title:       "Tesla rallies on Q1 results",
				Snippet:     "Tesla shares climbed after the company reported quarterly results above analyst expectations on revenue and margin.",
				URL:         "https://reuters.com/articles/tsla-q1",
				Source:      "reuters.com",
				PublishedAt: "2026-04-29",
				Score:       0.92,
			},
			{
				Title:       "TSLA delivery numbers",
				Snippet:     "Bloomberg report on TSLA delivery numbers indicates regional growth in Q1 across multiple geographies.",
				URL:         "https://bloomberg.com/news/tsla-deliveries",
				Source:      "bloomberg.com",
				PublishedAt: "2026-04-28",
				Score:       0.88,
			},
		},
	}
	cleanup := websearch.SetDefaultProviderForTest(mock)
	t.Cleanup(cleanup)

	in := service.BuildEnvelopeInputForChat("What is the latest news on Tesla today?")
	if strings.TrimSpace(in.WebContextBody) == "" {
		t.Fatalf("expected web context body, got empty")
	}
	if !strings.Contains(in.WebContextBody, "Tesla rallies on Q1 results") {
		t.Fatalf("expected first headline in body, got %q", in.WebContextBody)
	}
	if !strings.Contains(in.WebContextBody, "reuters.com") {
		t.Fatalf("expected source attribution, got %q", in.WebContextBody)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 provider call, got %d", len(mock.Calls))
	}
	if !strings.Contains(mock.Calls[0].Query, "TSLA") {
		t.Fatalf("expected TSLA in provider query, got %q", mock.Calls[0].Query)
	}

	full := chatprompt.BuildUserContent(in)
	if !strings.Contains(full, chatprompt.SectionWebContext) {
		t.Fatalf("rendered prompt missing %s section", chatprompt.SectionWebContext)
	}
	if !strings.Contains(full, "Tesla rallies on Q1 results") {
		t.Fatal("rendered prompt missing first headline")
	}
}

func TestChatEnvelope_EvergreenQuestionDoesNotCallProvider(t *testing.T) {
	isolateQA(t)

	mock := &websearch.MockProvider{
		Results: []websearch.Result{
			{Title: "should not appear", Snippet: "Long enough snippet that would otherwise pass.", URL: "https://reuters.com/x", Source: "reuters.com"},
		},
	}
	cleanup := websearch.SetDefaultProviderForTest(mock)
	t.Cleanup(cleanup)

	in := service.BuildEnvelopeInputForChat("What is beta in finance?")
	if strings.TrimSpace(in.WebContextBody) != "" {
		t.Fatalf("evergreen query should not attach web context, got %q", in.WebContextBody)
	}
	if len(mock.Calls) != 0 {
		t.Fatalf("expected 0 provider calls, got %d", len(mock.Calls))
	}

	full := chatprompt.BuildUserContent(in)
	if !strings.Contains(full, "No live web search results were attached") {
		t.Fatalf("rendered prompt should include empty-web-context note: %s", full)
	}
}

func TestChatEnvelope_ProviderErrorDegradesSilently(t *testing.T) {
	isolateQA(t)

	cleanup := websearch.SetDefaultProviderForTest(failingProvider{})
	t.Cleanup(cleanup)

	in := service.BuildEnvelopeInputForChat("Any news on Apple today?")
	if strings.TrimSpace(in.WebContextBody) != "" {
		t.Fatalf("provider error should yield empty web context, got %q", in.WebContextBody)
	}

	full := chatprompt.BuildUserContent(in)
	if !strings.Contains(full, "No live web search results were attached") {
		t.Fatalf("rendered prompt should include empty-web-context note: %s", full)
	}
}

func TestChatEnvelope_ProviderReturnsNoUsefulResults(t *testing.T) {
	isolateQA(t)

	mock := &websearch.MockProvider{
		Results: []websearch.Result{
			{Title: "", Snippet: "no title"},
			{Title: "tiny", Snippet: "x"},
		},
	}
	cleanup := websearch.SetDefaultProviderForTest(mock)
	t.Cleanup(cleanup)

	in := service.BuildEnvelopeInputForChat("Latest news on the market today")
	if strings.TrimSpace(in.WebContextBody) != "" {
		t.Fatalf("expected empty body after cleaning, got %q", in.WebContextBody)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("provider should still be called once, got %d", len(mock.Calls))
	}
}

func TestChatEnvelope_StubProviderProducesNoContext(t *testing.T) {
	isolateQA(t)

	cleanup := websearch.SetDefaultProviderForTest(websearch.StubProvider{})
	t.Cleanup(cleanup)

	in := service.BuildEnvelopeInputForChat("What's happening with Nvidia today?")
	if strings.TrimSpace(in.WebContextBody) != "" {
		t.Fatalf("stub provider must yield empty body, got %q", in.WebContextBody)
	}
}

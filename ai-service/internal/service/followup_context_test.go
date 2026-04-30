package service

import (
	"testing"

	"wealthscope-ai/internal/openai"
)

func TestResolveFollowUpContext_WithItCarriesTicker(t *testing.T) {
	st := openai.NewStore(openai.DefaultStoreConfig())
	prev := openai.SessionMessages("s-followup-it")
	_ = prev // silence linter if optimized away in future edits
	openai.SetDefaultStore(st)
	t.Cleanup(func() { openai.SetDefaultStore(openai.NewStore(openai.DefaultStoreConfig())) })

	st.AddUserMessage("s-followup-it", "What do you think about AAPL?")
	st.AddAssistantMessage("s-followup-it", "AAPL has shown strong earnings momentum.")

	ctx := resolveFollowUpContext("s-followup-it", "what about its risk?")
	if ctx.CarriedTicker != "AAPL" {
		t.Fatalf("expected AAPL carryover, got %+v", ctx)
	}
}

func TestApplyFollowUpCarryover_Compare(t *testing.T) {
	out := applyFollowUpCarryover("compare it with Microsoft", followUpContext{CarriedTicker: "AAPL"})
	if out == "compare it with Microsoft" {
		t.Fatalf("expected rewrite, got %q", out)
	}
	if out == "" {
		t.Fatal("rewritten message should not be empty")
	}
}

func TestResolveFollowUpContext_NewsFollowUp(t *testing.T) {
	st := openai.NewStore(openai.DefaultStoreConfig())
	openai.SetDefaultStore(st)
	t.Cleanup(func() { openai.SetDefaultStore(openai.NewStore(openai.DefaultStoreConfig())) })

	st.AddUserMessage("s-followup-news", "Tell me about TSLA")
	st.AddAssistantMessage("s-followup-news", "TSLA is in the EV sector.")

	ctx := resolveFollowUpContext("s-followup-news", "show me the latest news too")
	if ctx.CarriedTicker != "TSLA" {
		t.Fatalf("expected TSLA carryover, got %+v", ctx)
	}
}

func TestResolveFollowUpContext_PortfolioFollowUp(t *testing.T) {
	st := openai.NewStore(openai.DefaultStoreConfig())
	openai.SetDefaultStore(st)
	t.Cleanup(func() { openai.SetDefaultStore(openai.NewStore(openai.DefaultStoreConfig())) })

	st.AddUserMessage("s-followup-portfolio", "How risky is NVDA right now?")
	st.AddAssistantMessage("s-followup-portfolio", "NVDA can be relatively volatile.")

	ctx := resolveFollowUpContext("s-followup-portfolio", "how does that affect my portfolio?")
	if ctx.CarriedTicker != "NVDA" {
		t.Fatalf("expected NVDA carryover, got %+v", ctx)
	}
}

func TestResolveFollowUpContext_NoPriorContext(t *testing.T) {
	st := openai.NewStore(openai.DefaultStoreConfig())
	openai.SetDefaultStore(st)
	t.Cleanup(func() { openai.SetDefaultStore(openai.NewStore(openai.DefaultStoreConfig())) })

	ctx := resolveFollowUpContext("s-followup-empty", "what about its risk?")
	if ctx.CarriedTicker != "" {
		t.Fatalf("expected empty context, got %+v", ctx)
	}
}


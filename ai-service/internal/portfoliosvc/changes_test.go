package portfoliosvc

import (
	"testing"

	"wealthscope-ai/internal/ml"
)

func TestDescribeChanges_NoPrior(t *testing.T) {
	var req ChangesRequest
	req.Current.Holdings = []ml.PortfolioHolding{
		{Symbol: "AAPL", Allocation: 1, Beta: "1.2"},
	}
	resp, err := DescribeChanges(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.HasPriorSnapshot {
		t.Fatal("expected no prior")
	}
	if resp.ChangesSummary == "" {
		t.Fatal("expected summary")
	}
}

func TestDescribeChanges_WithPrior(t *testing.T) {
	var req ChangesRequest
	req.Current.Holdings = []ml.PortfolioHolding{
		{Symbol: "AAPL", Allocation: 0.5, Beta: "1.2"},
		{Symbol: "MSFT", Allocation: 0.5, Beta: "1.0"},
	}
	req.Prior.Holdings = []ml.PortfolioHolding{
		{Symbol: "AAPL", Allocation: 1, Beta: "1.2"},
	}
	resp, err := DescribeChanges(req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasPriorSnapshot {
		t.Fatal("expected prior")
	}
	if len(resp.SymbolsAdded) == 0 {
		t.Fatal("expected MSFT added")
	}
	if len(resp.SymbolsRemoved) != 0 {
		t.Fatalf("unexpected removed %v", resp.SymbolsRemoved)
	}
}

func TestDescribeChanges_EmptyCurrent(t *testing.T) {
	var req ChangesRequest
	req.Prior.Holdings = []ml.PortfolioHolding{{Symbol: "X", Allocation: 1, Beta: "1"}}
	_, err := DescribeChanges(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

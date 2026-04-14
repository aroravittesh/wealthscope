package portfoliosvc

import (
	"testing"

	"wealthscope-ai/internal/ml"
)

func TestSummarize_Basic(t *testing.T) {
	resp, err := Summarize(SummarizeRequest{
		Holdings: []ml.PortfolioHolding{
			{Symbol: "A", Allocation: 0.5, Beta: "1.0"},
			{Symbol: "B", Allocation: 0.5, Beta: "1.0"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PositionCount != 2 || resp.RiskLevel == "" {
		t.Fatalf("unexpected %+v", resp)
	}
	if len(resp.LargestPositions) < 2 {
		t.Fatal("expected largest positions")
	}
	if resp.Summary == "" || resp.Disclaimer == "" {
		t.Fatal("summary/disclaimer required")
	}
}

func TestSummarize_EmptyHoldings(t *testing.T) {
	_, err := Summarize(SummarizeRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

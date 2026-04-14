package tests

import (
	"strings"
	"testing"

	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/portfolioexplain"
)

func TestPortfolioExplain_HighRiskConcentrated(t *testing.T) {
	req := portfolioexplain.Request{
		TargetRisk: "LOW",
		Holdings: []ml.PortfolioHolding{
			{Symbol: "TSLA", Allocation: 0.5, Beta: "2.0"},
			{Symbol: "NVDA", Allocation: 0.5, Beta: "1.8"},
		},
	}
	resp, err := portfolioexplain.Explain(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Summary == "" || !strings.Contains(resp.Summary, "weighted beta") {
		t.Fatalf("summary: %q", resp.Summary)
	}
	if len(resp.TopRisks) == 0 {
		t.Fatal("expected top risks")
	}
	if !strings.Contains(strings.ToLower(resp.ConcentrationWarning), "concentration") {
		t.Fatalf("concentration_warning: %q", resp.ConcentrationWarning)
	}
	if resp.DiversificationNote == "" {
		t.Fatal("diversification note")
	}
	if !strings.Contains(resp.RiskAlignment, "target") && !strings.Contains(resp.RiskAlignment, "LOW") {
		t.Fatalf("risk_alignment: %q", resp.RiskAlignment)
	}
	if len(resp.NeutralGuidance) < 2 {
		t.Fatal("expected neutral guidance bullets")
	}
	forbidden := []string{"buy", "sell", "Buy", "Sell"}
	joined := resp.Summary + strings.Join(resp.TopRisks, " ") + strings.Join(resp.NeutralGuidance, " ")
	for _, w := range forbidden {
		if strings.Contains(joined, w) {
			t.Fatalf("output should avoid %q", w)
		}
	}
}

func TestPortfolioExplain_Balanced(t *testing.T) {
	req := portfolioexplain.Request{
		TargetRisk: "MEDIUM",
		Holdings: []ml.PortfolioHolding{
			{Symbol: "A", Allocation: 0.5, Beta: "1.0"},
			{Symbol: "B", Allocation: 0.5, Beta: "1.0"},
		},
	}
	resp, err := portfolioexplain.Explain(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Summary == "" {
		t.Fatal("empty summary")
	}
	if strings.Contains(resp.Summary, "sell") || strings.Contains(resp.Summary, "buy") {
		t.Fatal("no trade language")
	}
}

func TestPortfolioExplain_InvalidEmptyHoldings(t *testing.T) {
	_, err := portfolioexplain.Explain(portfolioexplain.Request{
		TargetRisk: "MEDIUM",
		Holdings:   nil,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPortfolioExplain_InvalidTarget(t *testing.T) {
	_, err := portfolioexplain.Explain(portfolioexplain.Request{
		TargetRisk: "INVALID",
		Holdings: []ml.PortfolioHolding{
			{Symbol: "X", Allocation: 1, Beta: "1"},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

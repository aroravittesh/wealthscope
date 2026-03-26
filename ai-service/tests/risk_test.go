package tests

import (
    "testing"
    "wealthscope-ai/internal/ml"
)

func TestRisk_HighRisk(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "TSLA", Allocation: 0.5, Beta: "2.0"},
        {Symbol: "NVDA", Allocation: 0.5, Beta: "1.8"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Level != ml.RiskHigh {
        t.Fatalf("expected HIGH got %s", report.Level)
    }
}

func TestRisk_MediumRisk(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "AAPL", Allocation: 0.5, Beta: "1.2"},
        {Symbol: "MSFT", Allocation: 0.5, Beta: "0.9"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Level != ml.RiskMedium {
        t.Fatalf("expected MEDIUM got %s", report.Level)
    }
}

func TestRisk_LowRisk(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "JNJ", Allocation: 0.5, Beta: "0.5"},
        {Symbol: "KO",  Allocation: 0.5, Beta: "0.6"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Level != ml.RiskLow {
        t.Fatalf("expected LOW got %s", report.Level)
    }
}

func TestRisk_InvalidBetaDefaultsToOne(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "AAPL", Allocation: 1.0, Beta: "invalid"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Score != 1.0 {
        t.Fatalf("expected score 1.0 got %f", report.Score)
    }
}

func TestRisk_ExplanationNotEmpty(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "AAPL", Allocation: 1.0, Beta: "1.2"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Explanation == "" {
        t.Fatal("expected non-empty explanation")
    }
}

func TestRisk_ScoreCalculation(t *testing.T) {
    holdings := []ml.PortfolioHolding{
        {Symbol: "AAPL", Allocation: 0.5, Beta: "1.0"},
        {Symbol: "MSFT", Allocation: 0.5, Beta: "1.0"},
    }
    report := ml.ScorePortfolio(holdings)
    if report.Score != 1.0 {
        t.Fatalf("expected score 1.0 got %f", report.Score)
    }
}
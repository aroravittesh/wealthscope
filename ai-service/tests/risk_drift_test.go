package tests

import (
	"testing"

	"wealthscope-ai/internal/prediction"
)

func TestPredictRiskDrift_LowTargetHighBeta(t *testing.T) {
	req := prediction.DriftRequest{
		TargetRisk: "LOW",
		Holdings: []prediction.DriftHolding{
			{Symbol: "TSLA", Allocation: 0.5, Beta: "2.0", Sector: "Consumer Cyclical"},
			{Symbol: "NVDA", Allocation: 0.5, Beta: "1.8", Sector: "Technology"},
		},
	}
	resp, err := prediction.PredictRiskDrift(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.DriftLevel != string(prediction.DriftHigh) {
		t.Fatalf("expected HIGH_DRIFT for aggressive book vs LOW target, got %s", resp.DriftLevel)
	}
	if resp.Score < 0.5 {
		t.Fatalf("expected elevated score, got %f", resp.Score)
	}
	if resp.Explanation == "" || resp.Disclaimer == "" {
		t.Fatal("expected explanation and disclaimer")
	}
}

func TestPredictRiskDrift_AlignedMedium(t *testing.T) {
	req := prediction.DriftRequest{
		TargetRisk: "MEDIUM",
		Holdings: []prediction.DriftHolding{
			{Symbol: "A", Allocation: 0.5, Beta: "1.0", Sector: "X"},
			{Symbol: "B", Allocation: 0.5, Beta: "1.0", Sector: "Y"},
		},
	}
	resp, err := prediction.PredictRiskDrift(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.DriftLevel == string(prediction.DriftHigh) {
		t.Fatalf("did not expect HIGH_DRIFT for balanced medium book, got score %f", resp.Score)
	}
}

func TestPredictRiskDrift_NormalizesWeights(t *testing.T) {
	req := prediction.DriftRequest{
		TargetRisk: "MEDIUM",
		Holdings: []prediction.DriftHolding{
			{Symbol: "AAPL", Allocation: 40, Beta: "1.0", Sector: "Technology"},
			{Symbol: "MSFT", Allocation: 60, Beta: "1.0", Sector: "Technology"},
		},
	}
	resp, err := prediction.PredictRiskDrift(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Score < 0 || resp.Score > 1 {
		t.Fatalf("score out of range: %f", resp.Score)
	}
}

func TestPredictRiskDrift_InvalidTarget(t *testing.T) {
	_, err := prediction.PredictRiskDrift(prediction.DriftRequest{
		TargetRisk: "BANANA",
		Holdings: []prediction.DriftHolding{
			{Symbol: "A", Allocation: 1, Beta: "1"},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

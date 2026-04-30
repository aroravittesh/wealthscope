package tests

import (
	"math"
	"testing"

	"wealthscope-ai/internal/portfoliorisk"
)

// --- Concentrated, high-risk portfolio: 70/30 TSLA/NVDA, single sector. ---

func TestPortfolioRisk_ConcentratedHighRisk(t *testing.T) {
	a := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "TSLA", Allocation: 0.70, Beta: "2.0", Sector: "Technology"},
		{Symbol: "NVDA", Allocation: 0.30, Beta: "1.8", Sector: "Technology"},
	}, "LOW")

	if a.RiskLevel != portfoliorisk.RiskBandHigh {
		t.Fatalf("expected HIGH risk for concentrated mix, got %s (score=%.3f)", a.RiskLevel, a.RiskScore)
	}
	if a.DriftLevel != portfoliorisk.DriftBandHigh {
		t.Fatalf("expected HIGH_DRIFT vs LOW target, got %s (score=%.3f)", a.DriftLevel, a.DriftScore)
	}
	if a.Features.PositionHHI < 0.55 {
		t.Fatalf("expected high position HHI, got %.3f", a.Features.PositionHHI)
	}
	if a.DiversificationScore > 0.45 {
		t.Fatalf("expected low diversification score, got %.3f", a.DiversificationScore)
	}
	if !hasDriver(a.TopRiskDrivers, "POSITION_CONCENTRATION") {
		t.Fatalf("expected POSITION_CONCENTRATION in top risk drivers, got %v", driverCodes(a.TopRiskDrivers))
	}
}

// --- Diversified medium-risk portfolio: skewed but spread across sectors. ---

func TestPortfolioRisk_DiversifiedMediumRisk(t *testing.T) {
	a := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "TSLA", Allocation: 0.40, Beta: "2.0", Sector: "Consumer Cyclical"},
		{Symbol: "AAPL", Allocation: 0.30, Beta: "1.2", Sector: "Technology"},
		{Symbol: "MSFT", Allocation: 0.20, Beta: "0.95", Sector: "Technology"},
		{Symbol: "JPM", Allocation: 0.10, Beta: "1.4", Sector: "Financial Services"},
	}, "MEDIUM")

	if a.RiskLevel != portfoliorisk.RiskBandMedium {
		t.Fatalf("expected MEDIUM risk, got %s (score=%.3f)", a.RiskLevel, a.RiskScore)
	}
	if a.DriftLevel == portfoliorisk.DriftBandHigh {
		t.Fatalf("did not expect HIGH_DRIFT for moderate book, got score=%.3f", a.DriftScore)
	}
	if a.Features.SectorCount < 3 {
		t.Fatalf("expected multi-sector book, got %d sectors", a.Features.SectorCount)
	}
}

// --- Defensive low-beta portfolio: spread across staples / healthcare. ---

func TestPortfolioRisk_DefensiveLowRisk(t *testing.T) {
	a := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "JNJ", Allocation: 0.20, Beta: "0.55", Sector: "Healthcare"},
		{Symbol: "KO", Allocation: 0.20, Beta: "0.60", Sector: "Consumer Staples"},
		{Symbol: "PG", Allocation: 0.20, Beta: "0.50", Sector: "Consumer Staples"},
		{Symbol: "VZ", Allocation: 0.20, Beta: "0.50", Sector: "Telecom"},
		{Symbol: "WMT", Allocation: 0.20, Beta: "0.50", Sector: "Consumer Staples"},
	}, "LOW")

	if a.RiskLevel != portfoliorisk.RiskBandLow {
		t.Fatalf("expected LOW risk for defensive mix, got %s (score=%.3f)", a.RiskLevel, a.RiskScore)
	}
	if a.Features.WeightedBeta > 0.7 {
		t.Fatalf("expected low weighted beta, got %.3f", a.Features.WeightedBeta)
	}
	if a.DriftLevel == portfoliorisk.DriftBandHigh {
		t.Fatalf("expected aligned drift for LOW target + low-beta book, got %s", a.DriftLevel)
	}
}

// --- Invalid inputs. ---

func TestPortfolioRisk_InvalidEmpty(t *testing.T) {
	if _, err := portfoliorisk.Assess(nil, "MEDIUM"); err == nil {
		t.Fatal("expected error for empty holdings")
	}
}

func TestPortfolioRisk_InvalidNegativeAllocation(t *testing.T) {
	_, err := portfoliorisk.Assess([]portfoliorisk.Holding{
		{Symbol: "AAPL", Allocation: -0.5, Beta: "1.1"},
	}, "MEDIUM")
	if err == nil {
		t.Fatal("expected error for negative allocation")
	}
}

func TestPortfolioRisk_InvalidZeroSum(t *testing.T) {
	_, err := portfoliorisk.Assess([]portfoliorisk.Holding{
		{Symbol: "AAPL", Allocation: 0, Beta: "1"},
		{Symbol: "MSFT", Allocation: 0, Beta: "1"},
	}, "MEDIUM")
	if err == nil {
		t.Fatal("expected error for zero-sum allocations")
	}
}

func TestPortfolioRisk_InvalidTarget(t *testing.T) {
	_, err := portfoliorisk.Assess([]portfoliorisk.Holding{
		{Symbol: "X", Allocation: 1, Beta: "1"},
	}, "BANANA")
	if err == nil {
		t.Fatal("expected error for unknown target")
	}
}

// --- Feature normalization: same shape regardless of allocation magnitude. ---

func TestPortfolioRisk_NormalizationStable(t *testing.T) {
	small, err1 := portfoliorisk.Assess([]portfoliorisk.Holding{
		{Symbol: "A", Allocation: 0.4, Beta: "1.0", Sector: "Tech"},
		{Symbol: "B", Allocation: 0.6, Beta: "1.0", Sector: "Tech"},
	}, "MEDIUM")
	big, err2 := portfoliorisk.Assess([]portfoliorisk.Holding{
		{Symbol: "A", Allocation: 4000, Beta: "1.0", Sector: "Tech"},
		{Symbol: "B", Allocation: 6000, Beta: "1.0", Sector: "Tech"},
	}, "MEDIUM")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if !floatNear(small.RiskScore, big.RiskScore, 1e-6) ||
		!floatNear(small.DriftScore, big.DriftScore, 1e-6) ||
		!floatNear(small.Features.WeightedBeta, big.Features.WeightedBeta, 1e-6) ||
		!floatNear(small.Features.PositionHHI, big.Features.PositionHHI, 1e-6) {
		t.Fatalf("normalisation mismatch:\n small=%+v\n big=%+v", small, big)
	}
}

// --- Confidence drops when sector / volatility data is missing. ---

func TestPortfolioRisk_DataQualityAffectsConfidence(t *testing.T) {
	full := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 0.34, Beta: "1.1", Sector: "Tech", Volatility24h: "1.5"},
		{Symbol: "B", Allocation: 0.33, Beta: "0.95", Sector: "Healthcare", Volatility24h: "1.2"},
		{Symbol: "C", Allocation: 0.33, Beta: "1.0", Sector: "Financial", Volatility24h: "1.0"},
	}, "MEDIUM")
	sparse := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 0.5, Beta: "1.1"},
		{Symbol: "B", Allocation: 0.5, Beta: "0.95"},
	}, "MEDIUM")
	if full.Confidence <= sparse.Confidence {
		t.Fatalf("expected richer data to yield higher confidence: full=%.3f sparse=%.3f", full.Confidence, sparse.Confidence)
	}
}

// --- Sentiment data should pull a bullish book down on the bearish term. ---

func TestPortfolioRisk_SentimentLiftsRiskWhenBearish(t *testing.T) {
	bullish := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 0.5, Beta: "1.0", Sector: "Tech", Sentiment: "BULLISH"},
		{Symbol: "B", Allocation: 0.5, Beta: "1.0", Sector: "Healthcare", Sentiment: "BULLISH"},
	}, "")
	bearish := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 0.5, Beta: "1.0", Sector: "Tech", Sentiment: "BEARISH"},
		{Symbol: "B", Allocation: 0.5, Beta: "1.0", Sector: "Healthcare", Sentiment: "BEARISH"},
	}, "")
	if bearish.RiskScore <= bullish.RiskScore {
		t.Fatalf("expected bearish news to raise risk: bullish=%.3f bearish=%.3f", bullish.RiskScore, bearish.RiskScore)
	}
	if !hasDriver(bearish.TopRiskDrivers, "SENTIMENT") {
		t.Fatalf("expected SENTIMENT driver in bearish case, got %v", driverCodes(bearish.TopRiskDrivers))
	}
}

// --- Top drivers reflect dominant features (concentration here). ---

func TestPortfolioRisk_TopDriversIdentifyConcentration(t *testing.T) {
	a := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "AAPL", Allocation: 0.85, Beta: "1.2", Sector: "Technology"},
		{Symbol: "MSFT", Allocation: 0.15, Beta: "1.1", Sector: "Technology"},
	}, "MEDIUM")
	if !hasDriver(a.TopRiskDrivers, "POSITION_CONCENTRATION") {
		t.Fatalf("expected POSITION_CONCENTRATION as risk driver, got %v", driverCodes(a.TopRiskDrivers))
	}
	if !hasDriver(a.TopRiskDrivers, "SECTOR_CONCENTRATION") {
		t.Fatalf("expected SECTOR_CONCENTRATION as risk driver, got %v", driverCodes(a.TopRiskDrivers))
	}
	if a.Features.EffectiveN >= 2 {
		t.Fatalf("expected effective_n < 2 for highly skewed mix, got %.3f", a.Features.EffectiveN)
	}
}

// --- Drift increases as the weighted beta drifts from the target center. ---

func TestPortfolioRisk_DriftMonotonicVsTarget(t *testing.T) {
	low := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 1.0, Beta: "1.0", Sector: "Tech"},
	}, "MEDIUM")
	mid := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 1.0, Beta: "1.5", Sector: "Tech"},
	}, "MEDIUM")
	high := mustAssess(t, []portfoliorisk.Holding{
		{Symbol: "A", Allocation: 1.0, Beta: "2.5", Sector: "Tech"},
	}, "MEDIUM")
	if !(low.DriftScore < mid.DriftScore && mid.DriftScore < high.DriftScore) {
		t.Fatalf("expected drift to be monotonic vs beta gap: %.3f / %.3f / %.3f",
			low.DriftScore, mid.DriftScore, high.DriftScore)
	}
}

// --- Helpers ---

func mustAssess(t *testing.T, hs []portfoliorisk.Holding, target string) portfoliorisk.Assessment {
	t.Helper()
	a, err := portfoliorisk.Assess(hs, target)
	if err != nil {
		t.Fatalf("assess error: %v", err)
	}
	if a.RiskScore < 0 || a.RiskScore > 1 {
		t.Fatalf("risk score out of range: %.4f", a.RiskScore)
	}
	if target != "" && (a.DriftScore < 0 || a.DriftScore > 1) {
		t.Fatalf("drift score out of range: %.4f", a.DriftScore)
	}
	if a.Confidence < 0 || a.Confidence > 1 {
		t.Fatalf("confidence out of range: %.4f", a.Confidence)
	}
	return a
}

func hasDriver(drivers []portfoliorisk.Driver, code string) bool {
	for _, d := range drivers {
		if d.Code == code {
			return true
		}
	}
	return false
}

func driverCodes(drivers []portfoliorisk.Driver) []string {
	out := make([]string, len(drivers))
	for i, d := range drivers {
		out[i] = d.Code
	}
	return out
}

func floatNear(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

package ml

import (
	"fmt"
	"strconv"

	"wealthscope-ai/internal/explain"
	"wealthscope-ai/internal/portfoliorisk"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)

type PortfolioHolding struct {
	Symbol     string  `json:"symbol"`
	Allocation float64 `json:"allocation"`
	Beta       string  `json:"beta"`
}

// RiskReport keeps the legacy beta-based shape (`Score`, `Level`,
// `Explanation`) for backward compatibility, and adds optional richer fields
// produced by the upgraded portfoliorisk pipeline.
type RiskReport struct {
	Score       float64   `json:"Score"`
	Level       RiskLevel `json:"Level"`
	Explanation string    `json:"Explanation"`

	// Enriched additive fields (populated when the upgraded pipeline runs
	// without errors). All are tagged omitempty to remain non-breaking.
	WeightedBeta         float64                `json:"weighted_beta,omitempty"`
	PositionHHI          float64                `json:"position_hhi,omitempty"`
	EffectiveN           float64                `json:"effective_n,omitempty"`
	DiversificationScore float64                `json:"diversification_score,omitempty"`
	CompositeRiskScore   float64                `json:"composite_risk_score,omitempty"`
	CompositeRiskLevel   string                 `json:"composite_risk_level,omitempty"`
	Confidence           float64                `json:"confidence,omitempty"`
	TopDrivers           []portfoliorisk.Driver `json:"top_drivers,omitempty"`

	// Explainability envelope (additive; safe for older clients to ignore).
	// The legacy free-text `Explanation` field above is unchanged.
	ReasonCode        string               `json:"reason_code,omitempty"`
	ReasoningSummary  string               `json:"reasoning_summary,omitempty"`
	ExplanationDetail *explain.Explanation `json:"explanation_detail,omitempty"`
}

// ScorePortfolio computes the legacy beta-based score and overlays the
// richer portfoliorisk pipeline output. The legacy `Score`, `Level`, and
// `Explanation` fields keep their previous semantics so existing clients and
// tests are unaffected.
func ScorePortfolio(holdings []PortfolioHolding) RiskReport {
	weightedBeta := legacyWeightedBeta(holdings)

	level := legacyLevelFromBeta(weightedBeta)
	explanation := legacyExplanation(weightedBeta, level)

	report := RiskReport{
		Score:       weightedBeta,
		Level:       level,
		Explanation: explanation,
	}

	in := make([]portfoliorisk.Holding, len(holdings))
	for i, h := range holdings {
		in[i] = portfoliorisk.Holding{
			Symbol:     h.Symbol,
			Allocation: h.Allocation,
			Beta:       h.Beta,
		}
	}
	if a, err := portfoliorisk.Assess(in, ""); err == nil {
		report.WeightedBeta = a.Features.WeightedBeta
		report.PositionHHI = a.Features.PositionHHI
		report.EffectiveN = a.Features.EffectiveN
		report.DiversificationScore = a.DiversificationScore
		report.CompositeRiskScore = a.RiskScore
		report.CompositeRiskLevel = string(a.RiskLevel)
		report.Confidence = a.Confidence
		report.TopDrivers = a.TopRiskDrivers

		exp := explain.BuildRiskExplanation(explain.RiskInputs{
			Level:      string(a.RiskLevel),
			Score:      a.RiskScore,
			Confidence: a.Confidence,
			Drivers:    a.TopRiskDrivers,
		})
		report.ReasonCode = exp.Code
		report.ReasoningSummary = exp.Summary
		report.ExplanationDetail = &exp
	}

	return report
}

func legacyWeightedBeta(holdings []PortfolioHolding) float64 {
	wBeta := 0.0
	for _, h := range holdings {
		beta, err := strconv.ParseFloat(h.Beta, 64)
		if err != nil {
			beta = 1.0
		}
		wBeta += beta * h.Allocation
	}
	return wBeta
}

func legacyLevelFromBeta(beta float64) RiskLevel {
	switch {
	case beta >= 1.5:
		return RiskHigh
	case beta >= 0.8:
		return RiskMedium
	default:
		return RiskLow
	}
}

func legacyExplanation(beta float64, level RiskLevel) string {
	switch level {
	case RiskHigh:
		return fmt.Sprintf(
			"Your portfolio has a weighted beta of %.2f, indicating high volatility relative to the market. Consider adding stable, low-beta assets.",
			beta,
		)
	case RiskMedium:
		return fmt.Sprintf(
			"Your portfolio has a weighted beta of %.2f, broadly in line with market movements. A balanced profile.",
			beta,
		)
	default:
		return fmt.Sprintf(
			"Your portfolio has a weighted beta of %.2f, suggesting low volatility. May underperform in strong bull markets.",
			beta,
		)
	}
}

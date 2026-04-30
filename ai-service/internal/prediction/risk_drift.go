package prediction

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/explain"
	"wealthscope-ai/internal/portfoliorisk"
)

type DriftLevel string

const (
	DriftLow    DriftLevel = "LOW_DRIFT"
	DriftMedium DriftLevel = "MEDIUM_DRIFT"
	DriftHigh   DriftLevel = "HIGH_DRIFT"
)

// DriftHolding is one position for drift estimation. Sector / Volatility24 /
// Sentiment are optional; missing fields lower confidence rather than poison
// the score.
type DriftHolding struct {
	Symbol       string  `json:"symbol"`
	Allocation   float64 `json:"allocation"`
	Beta         string  `json:"beta"`
	Sector       string  `json:"sector,omitempty"`
	Volatility24 string  `json:"volatility_24h,omitempty"`
	Sentiment    string  `json:"sentiment,omitempty"`
}

// DriftRequest is the payload for POST /predict/risk-drift.
type DriftRequest struct {
	Holdings   []DriftHolding `json:"holdings"`
	TargetRisk string         `json:"target_risk"` // LOW, MEDIUM, HIGH
}

// DriftResponse keeps the original drift fields and adds richer optional
// outputs for the upgraded portfoliorisk pipeline. Older clients keep
// working unchanged because every new field is JSON omitempty.
type DriftResponse struct {
	DriftLevel  string  `json:"drift_level"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
	Disclaimer  string  `json:"disclaimer"`

	WeightedBeta float64 `json:"weighted_beta,omitempty"`
	PositionHHI  float64 `json:"position_hhi,omitempty"`
	SectorHHI    float64 `json:"sector_hhi,omitempty"`

	EffectiveN           float64                  `json:"effective_n,omitempty"`
	Top3Concentration    float64                  `json:"top3_concentration,omitempty"`
	AllocationGini       float64                  `json:"allocation_gini,omitempty"`
	WeightedVolatility   float64                  `json:"weighted_volatility,omitempty"`
	DrawdownProxy        float64                  `json:"drawdown_proxy,omitempty"`
	CorrelationProxy     float64                  `json:"correlation_proxy,omitempty"`
	DiversificationScore float64                  `json:"diversification_score,omitempty"`
	SentimentScore       float64                  `json:"sentiment_score,omitempty"`
	Confidence           float64                  `json:"confidence,omitempty"`
	TargetCenterBeta     float64                  `json:"target_center_beta,omitempty"`
	TargetMisalignment   float64                  `json:"target_misalignment,omitempty"`
	Components           map[string]float64       `json:"components,omitempty"`
	TopDrivers           []portfoliorisk.Driver   `json:"top_drivers,omitempty"`
	RiskScore            float64                  `json:"risk_score,omitempty"`
	RiskLevel            string                   `json:"risk_level,omitempty"`

	// Explainability envelope (additive; safe for older clients to ignore).
	// The legacy free-text `Explanation` field above is unchanged.
	ReasonCode        string               `json:"reason_code,omitempty"`
	ReasoningSummary  string               `json:"reasoning_summary,omitempty"`
	ExplanationDetail *explain.Explanation `json:"explanation_detail,omitempty"`
}

const disclaimer = portfoliorisk.Disclaimer

func PredictRiskDrift(req DriftRequest) (DriftResponse, error) {
	if len(req.Holdings) == 0 {
		return DriftResponse{}, fmt.Errorf("holdings required")
	}

	holdings := make([]portfoliorisk.Holding, 0, len(req.Holdings))
	for _, h := range req.Holdings {
		holdings = append(holdings, portfoliorisk.Holding{
			Symbol:        h.Symbol,
			Allocation:    h.Allocation,
			Beta:          h.Beta,
			Sector:        h.Sector,
			Volatility24h: h.Volatility24,
			Sentiment:     h.Sentiment,
		})
	}

	a, err := portfoliorisk.Assess(holdings, req.TargetRisk)
	if err != nil {
		return DriftResponse{}, err
	}

	target := strings.ToLower(strings.TrimSpace(req.TargetRisk))

	narrative := fmt.Sprintf(
		"Estimated portfolio beta is %.2f vs a %.2f reference for target %s. "+
			"Position concentration (HHI) is %.2f (1.0 = single stock). "+
			"Sector concentration HHI is %.2f. "+
			"Larger gaps and higher concentration raise drift vs your stated risk profile.",
		a.Features.WeightedBeta, a.TargetCenterBeta, target,
		a.Features.PositionHHI, a.Features.SectorHHI,
	)

	expEnv := explain.BuildDriftExplanation(explain.DriftInputs{
		Level:        a.DriftLevel,
		Score:        a.DriftScore,
		Confidence:   a.Confidence,
		Target:       req.TargetRisk,
		WeightedBeta: a.Features.WeightedBeta,
		CenterBeta:   a.TargetCenterBeta,
		Misalignment: a.TargetMisalignment,
		Drivers:      a.TopDriftDrivers,
	})

	return DriftResponse{
		DriftLevel:           a.DriftLevel,
		Score:                a.DriftScore,
		Explanation:          narrative,
		ReasonCode:           expEnv.Code,
		ReasoningSummary:     expEnv.Summary,
		ExplanationDetail:    &expEnv,
		Disclaimer:           disclaimer,
		WeightedBeta:         a.Features.WeightedBeta,
		PositionHHI:          a.Features.PositionHHI,
		SectorHHI:            a.Features.SectorHHI,
		EffectiveN:           a.Features.EffectiveN,
		Top3Concentration:    a.Features.Top3Concentration,
		AllocationGini:       a.Features.AllocationGini,
		WeightedVolatility:   a.Features.WeightedVolatility,
		DrawdownProxy:        a.Features.DrawdownProxy,
		CorrelationProxy:     a.Features.CorrelationProxy,
		DiversificationScore: a.DiversificationScore,
		SentimentScore:       a.Features.SentimentScore,
		Confidence:           a.Confidence,
		TargetCenterBeta:     a.TargetCenterBeta,
		TargetMisalignment:   a.TargetMisalignment,
		Components:           a.DriftComponents,
		TopDrivers:           a.TopDriftDrivers,
		RiskScore:            a.RiskScore,
		RiskLevel:            string(a.RiskLevel),
	}, nil
}

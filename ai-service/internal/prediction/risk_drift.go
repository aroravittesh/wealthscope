package prediction

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type DriftLevel string

const (
	DriftLow    DriftLevel = "LOW_DRIFT"
	DriftMedium DriftLevel = "MEDIUM_DRIFT"
	DriftHigh   DriftLevel = "HIGH_DRIFT"
)

// DriftHolding is one position for drift estimation (sector optional).
type DriftHolding struct {
	Symbol       string  `json:"symbol"`
	Allocation   float64 `json:"allocation"`
	Beta         string  `json:"beta"`
	Sector       string  `json:"sector,omitempty"`
	Volatility24 string  `json:"volatility_24h,omitempty"` // optional % string e.g. "2.5"
}

// DriftRequest is the payload for POST /predict/risk-drift.
type DriftRequest struct {
	Holdings   []DriftHolding `json:"holdings"`
	TargetRisk string         `json:"target_risk"` // LOW, MEDIUM, HIGH
}

// DriftResponse is a heuristic drift estimate, not investment advice.
type DriftResponse struct {
	DriftLevel  string  `json:"drift_level"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
	Disclaimer  string  `json:"disclaimer"`
}

const disclaimer = "This is a rough estimation for discussion and education only. It is not financial advice and does not predict future performance."

func PredictRiskDrift(req DriftRequest) (DriftResponse, error) {
	if len(req.Holdings) == 0 {
		return DriftResponse{}, fmt.Errorf("holdings required")
	}
	target := strings.ToUpper(strings.TrimSpace(req.TargetRisk))
	center, ok := targetBetaCenter(target)
	if !ok {
		return DriftResponse{}, fmt.Errorf("target_risk must be LOW, MEDIUM, or HIGH")
	}

	sumW := 0.0
	for _, h := range req.Holdings {
		if h.Allocation < 0 {
			return DriftResponse{}, fmt.Errorf("negative allocation for %s", h.Symbol)
		}
		sumW += h.Allocation
	}
	if sumW == 0 {
		return DriftResponse{}, fmt.Errorf("allocations must sum to a positive value")
	}
	invSum := 1.0 / sumW
	weightedBeta := 0.0
	hhi := 0.0
	sectorW := make(map[string]float64)
	volProxy := 0.0
	volWeight := 0.0

	for _, h := range req.Holdings {
		w := h.Allocation * invSum
		beta, err := strconv.ParseFloat(strings.TrimSpace(h.Beta), 64)
		if err != nil {
			beta = 1.0
		}
		weightedBeta += beta * w
		hhi += w * w

		sec := strings.TrimSpace(h.Sector)
		if sec == "" {
			sec = "UNKNOWN"
		}
		sectorW[sec] += w

		if v, err := strconv.ParseFloat(strings.TrimSpace(h.Volatility24), 64); err == nil && v >= 0 {
			volProxy += v * w
			volWeight += w
		}
	}

	sectorHHI := 0.0
	for _, sw := range sectorW {
		sectorHHI += sw * sw
	}

	betaGap := math.Abs(weightedBeta-center) / math.Max(0.15, center)
	betaTerm := math.Min(1.0, betaGap)

	concTerm := math.Min(1.0, math.Max(0, hhi-0.2)/0.6)
	sectorTerm := math.Min(1.0, math.Max(0, sectorHHI-0.2)/0.6)

	volTerm := 0.0
	if volWeight > 0 {
		v := volProxy / volWeight
		volTerm = math.Min(1.0, v/5.0)
	}

	raw := 0.55*betaTerm + 0.20*concTerm + 0.15*sectorTerm + 0.10*volTerm
	raw = math.Min(1.0, math.Max(0.0, raw))

	level := DriftLow
	switch {
	case raw >= 0.58:
		level = DriftHigh
	case raw >= 0.30:
		level = DriftMedium
	}

	explain := fmt.Sprintf(
		"Estimated portfolio beta is %.2f vs a %.2f reference for target %s. "+
			"Position concentration (HHI) is %.2f (1.0 = single stock). "+
			"Sector concentration HHI is %.2f. "+
			"Larger gaps and higher concentration raise drift vs your stated risk profile.",
		weightedBeta, center, strings.ToLower(target), hhi, sectorHHI,
	)

	return DriftResponse{
		DriftLevel:  string(level),
		Score:       raw,
		Explanation: explain,
		Disclaimer:  disclaimer,
	}, nil
}

func targetBetaCenter(target string) (float64, bool) {
	switch target {
	case "LOW":
		return 0.65, true
	case "MEDIUM":
		return 1.0, true
	case "HIGH":
		return 1.35, true
	default:
		return 0, false
	}
}

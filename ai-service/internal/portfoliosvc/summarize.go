package portfoliosvc

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/prediction"
)

// SummarizeRequest is the JSON body for POST /portfolio/summarize.
type SummarizeRequest struct {
	Holdings []ml.PortfolioHolding `json:"holdings"`
}

// PositionLine is one row in largest_positions.
type PositionLine struct {
	Symbol     string  `json:"symbol"`
	Allocation float64 `json:"allocation"`
	Beta       string  `json:"beta"`
	NormWeight float64 `json:"normalized_weight"`
}

// SummarizeResponse is the JSON output for POST /portfolio/summarize.
type SummarizeResponse struct {
	PositionCount            int            `json:"position_count"`
	WeightedBeta             float64        `json:"weighted_beta"`
	PositionConcentrationHHI float64        `json:"position_concentration_hhi"`
	SectorConcentrationHHI   float64        `json:"sector_concentration_hhi"`
	RiskLevel                string         `json:"risk_level"`
	LargestPositions         []PositionLine `json:"largest_positions"`
	Summary                  string         `json:"summary"`
	Disclaimer               string         `json:"disclaimer"`
}

// Summarize produces a concise structural summary of a holdings snapshot.
func Summarize(req SummarizeRequest) (*SummarizeResponse, error) {
	if len(req.Holdings) == 0 {
		return nil, fmt.Errorf("holdings required")
	}

	drift, err := prediction.PredictRiskDrift(prediction.DriftRequest{
		Holdings:   prediction.DriftHoldingsFromML(req.Holdings),
		TargetRisk: "MEDIUM",
	})
	if err != nil {
		return nil, err
	}

	report := ml.ScorePortfolio(normalizeHoldings(req.Holdings))
	lines := largestByWeight(req.Holdings, 5)

	summary := fmt.Sprintf(
		"The snapshot has %d position(s), estimated weighted beta %.2f, and position concentration (HHI) %.2f. "+
			"Sector concentration HHI is %.2f when sectors are known. "+
			"On a simple beta scale the mix reads as %s risk. "+
			"Largest weights are listed below for quick scanning.",
		len(req.Holdings),
		round4(drift.WeightedBeta),
		round4(drift.PositionHHI),
		round4(drift.SectorHHI),
		strings.ToLower(string(report.Level)),
	)

	return &SummarizeResponse{
		PositionCount:            len(req.Holdings),
		WeightedBeta:             round4(drift.WeightedBeta),
		PositionConcentrationHHI: round4(drift.PositionHHI),
		SectorConcentrationHHI:   round4(drift.SectorHHI),
		RiskLevel:                string(report.Level),
		LargestPositions:         lines,
		Summary:                  summary,
		Disclaimer:               EducationalDisclaimer,
	}, nil
}

func largestByWeight(h []ml.PortfolioHolding, k int) []PositionLine {
	sum := 0.0
	for _, x := range h {
		sum += x.Allocation
	}
	if sum <= 0 {
		return nil
	}
	inv := 1.0 / sum
	cp := append([]ml.PortfolioHolding(nil), h...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Allocation > cp[j].Allocation })
	if k > len(cp) {
		k = len(cp)
	}
	out := make([]PositionLine, 0, k)
	for i := 0; i < k; i++ {
		x := cp[i]
		out = append(out, PositionLine{
			Symbol:     strings.ToUpper(strings.TrimSpace(x.Symbol)),
			Allocation: round4(x.Allocation),
			Beta:       x.Beta,
			NormWeight: round4(x.Allocation * inv),
		})
	}
	return out
}

func round4(x float64) float64 {
	return math.Round(x*10000) / 10000
}

func normalizeHoldings(h []ml.PortfolioHolding) []ml.PortfolioHolding {
	sum := 0.0
	for _, x := range h {
		sum += x.Allocation
	}
	if sum <= 0 {
		return h
	}
	out := make([]ml.PortfolioHolding, len(h))
	for i, x := range h {
		out[i] = x
		out[i].Allocation = x.Allocation / sum
	}
	return out
}

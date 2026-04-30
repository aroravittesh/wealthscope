package portfoliorisk

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Driver is one explanation-friendly contributor to a score.
type Driver struct {
	Code         string  `json:"code"`
	Label        string  `json:"label"`
	Contribution float64 `json:"contribution"` // weighted contribution to the parent score
	Value        float64 `json:"value"`        // raw normalised feature value in [0, 1]
	Detail       string  `json:"detail"`
}

// RiskBand is a coarse risk bucket derived from the composite score.
type RiskBand string

const (
	RiskBandLow    RiskBand = "LOW"
	RiskBandMedium RiskBand = "MEDIUM"
	RiskBandHigh   RiskBand = "HIGH"
)

// DriftBand mirrors the legacy DriftLevel string values for compatibility.
const (
	DriftBandLow    = "LOW_DRIFT"
	DriftBandMedium = "MEDIUM_DRIFT"
	DriftBandHigh   = "HIGH_DRIFT"
)

// Assessment is the full output of the upgraded risk pipeline.
type Assessment struct {
	Features             Features           `json:"features"`
	RiskScore            float64            `json:"risk_score"`
	RiskLevel            RiskBand           `json:"risk_level"`
	DriftScore           float64            `json:"drift_score,omitempty"`
	DriftLevel           string             `json:"drift_level,omitempty"`
	DiversificationScore float64            `json:"diversification_score"`
	Confidence           float64            `json:"confidence"`
	TargetMisalignment   float64            `json:"target_misalignment,omitempty"`
	TargetCenterBeta     float64            `json:"target_center_beta,omitempty"`
	RiskComponents       map[string]float64 `json:"risk_components,omitempty"`
	DriftComponents      map[string]float64 `json:"drift_components,omitempty"`
	TopRiskDrivers       []Driver           `json:"top_risk_drivers,omitempty"`
	TopDriftDrivers      []Driver           `json:"top_drift_drivers,omitempty"`
}

// Risk weights (always sum to 1.0). Sentiment slot is redistributed when
// sentiment data isn't available so the score stays in [0, 1].
var riskWeights = map[string]float64{
	"BETA":                  0.30,
	"POSITION_CONCENTRATION": 0.20,
	"SECTOR_CONCENTRATION":  0.15,
	"VOLATILITY":            0.10,
	"ALLOCATION_IMBALANCE":  0.05,
	"DRAWDOWN":              0.05,
	"CORRELATION":           0.05,
	"SENTIMENT":             0.10,
}

// Drift weights anchored on misalignment-to-target.
var driftWeights = map[string]float64{
	"MISALIGNMENT":          0.45,
	"POSITION_CONCENTRATION": 0.20,
	"SECTOR_CONCENTRATION":  0.15,
	"VOLATILITY":            0.10,
	"ALLOCATION_IMBALANCE":  0.04,
	"DRAWDOWN":              0.03,
	"CORRELATION":           0.03,
}

// Risk band thresholds calibrated against representative books (concentrated
// 70/30 single-sector → ~0.50, balanced ~0.20, defensive multi-sector → <0.20).
const (
	riskHighThreshold   = 0.45
	riskMediumThreshold = 0.20

	driftHighThreshold   = 0.58
	driftMediumThreshold = 0.30
)

// Disclaimer is the standard educational footer.
const Disclaimer = "This is a rough estimation for discussion and education only. It is not financial advice and does not predict future performance."

// Assess runs the full risk + (optional) drift pipeline.
//
// targetRisk may be "" to skip drift scoring; otherwise "LOW" / "MEDIUM" /
// "HIGH". The function never panics on missing optional fields; instead the
// confidence drops to reflect partial data.
func Assess(holdings []Holding, targetRisk string) (Assessment, error) {
	feat, err := ExtractFeatures(holdings)
	if err != nil {
		return Assessment{}, err
	}

	risk, riskComps, riskDrivers := computeRisk(feat)

	a := Assessment{
		Features:             feat,
		RiskScore:            round6(risk),
		RiskLevel:            riskBand(risk),
		DiversificationScore: round6(diversificationScore(feat)),
		Confidence:           round6(confidence(feat)),
		RiskComponents:       roundComps(riskComps),
		TopRiskDrivers:       riskDrivers,
	}

	target := strings.ToUpper(strings.TrimSpace(targetRisk))
	if target != "" {
		center, ok := targetBetaCenter(target)
		if !ok {
			return Assessment{}, fmt.Errorf("target_risk must be LOW, MEDIUM, or HIGH")
		}
		drift, driftComps, driftDrivers, misalign := computeDrift(feat, center)
		a.DriftScore = round6(drift)
		a.DriftLevel = driftBand(drift)
		a.DriftComponents = roundComps(driftComps)
		a.TopDriftDrivers = driftDrivers
		a.TargetMisalignment = round6(misalign)
		a.TargetCenterBeta = center
	}

	return a, nil
}

// computeRisk fuses normalised feature terms into a single risk score in [0,1].
func computeRisk(f Features) (score float64, comps map[string]float64, drivers []Driver) {
	terms := termValues(f, 0)
	weights := make(map[string]float64, len(riskWeights))
	for k, v := range riskWeights {
		weights[k] = v
	}
	// If no sentiment coverage, redistribute sentiment weight proportionally
	// across the other risk components so the score stays in [0, 1].
	if f.SentimentCoverage <= 0 {
		share := weights["SENTIMENT"]
		delete(weights, "SENTIMENT")
		var rest float64
		for _, v := range weights {
			rest += v
		}
		for k, v := range weights {
			weights[k] = v + share*(v/rest)
		}
	}
	comps = make(map[string]float64, len(weights))
	for code, w := range weights {
		v := terms[code]
		c := w * v
		comps[code] = c
		score += c
	}
	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}
	drivers = topDrivers(comps, terms, 3)
	return score, comps, drivers
}

// computeDrift fuses the drift-anchored term set against a target beta center.
func computeDrift(f Features, center float64) (score float64, comps map[string]float64, drivers []Driver, misalign float64) {
	misalign = misalignmentTerm(f.WeightedBeta, center)
	terms := termValues(f, misalign)
	comps = make(map[string]float64, len(driftWeights))
	for code, w := range driftWeights {
		v := terms[code]
		c := w * v
		comps[code] = c
		score += c
	}
	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}
	drivers = topDrivers(comps, terms, 3)
	return score, comps, drivers, misalign
}

// termValues maps each component code to its [0, 1] term value.
// Missing-data terms collapse cleanly to 0 (handled by coverage gates).
func termValues(f Features, misalignment float64) map[string]float64 {
	betaTerm := clamp01(math.Abs(f.WeightedBeta-1.0) / 1.5)
	concTerm := clamp01(math.Max(0, f.PositionHHI-0.20) / 0.60)
	sectorTerm := 0.0
	if f.SectorCoverage > 0 {
		sectorTerm = clamp01(math.Max(0, f.SectorHHI-0.20) / 0.60)
	}
	volTerm := 0.0
	if f.VolatilityCoverage > 0 {
		volTerm = clamp01(f.WeightedVolatility / 5.0)
	}
	drawdownTerm := 0.0
	if f.VolatilityCoverage > 0 {
		drawdownTerm = clamp01(f.DrawdownProxy / 3.0)
	}
	corrTerm := 0.0
	if f.SectorCoverage > 0 {
		corrTerm = clamp01(f.CorrelationProxy / 0.5)
	}
	giniTerm := clamp01(f.AllocationGini)
	sentTerm := 0.0
	if f.SentimentCoverage > 0 {
		sentTerm = clamp01(f.SentimentScore)
	}

	return map[string]float64{
		"BETA":                  betaTerm,
		"MISALIGNMENT":          clamp01(misalignment),
		"POSITION_CONCENTRATION": concTerm,
		"SECTOR_CONCENTRATION":  sectorTerm,
		"VOLATILITY":            volTerm,
		"ALLOCATION_IMBALANCE":  giniTerm,
		"DRAWDOWN":              drawdownTerm,
		"CORRELATION":           corrTerm,
		"SENTIMENT":             sentTerm,
	}
}

func misalignmentTerm(weightedBeta, center float64) float64 {
	if center <= 0 {
		return 0
	}
	return clamp01(math.Abs(weightedBeta-center) / math.Max(0.15, center))
}

// diversificationScore is a "higher = better" composite useful for the UI.
func diversificationScore(f Features) float64 {
	posSpread := 1.0 - f.PositionHHI
	sectorSpread := 0.5 // neutral baseline when sector unknown
	if f.SectorCoverage > 0 {
		sectorSpread = 1.0 - f.SectorHHI
	}
	effSpread := math.Min(1.0, f.EffectiveN/8.0)
	giniSpread := 1.0 - f.AllocationGini
	d := 0.45*posSpread + 0.30*sectorSpread + 0.15*effSpread + 0.10*giniSpread
	return clamp01(d)
}

// confidence reflects data quality + position evidence + sector evidence.
func confidence(f Features) float64 {
	dataQ := (f.BetaCoverage + f.SectorCoverage + f.VolatilityCoverage) / 3.0
	posEv := math.Min(1.0, float64(f.NumPositions)/5.0)
	secEv := math.Min(1.0, float64(f.SectorCount)/3.0)
	return clamp01(0.5*dataQ + 0.3*posEv + 0.2*secEv)
}

func riskBand(score float64) RiskBand {
	switch {
	case score >= riskHighThreshold:
		return RiskBandHigh
	case score >= riskMediumThreshold:
		return RiskBandMedium
	default:
		return RiskBandLow
	}
}

func driftBand(score float64) string {
	switch {
	case score >= driftHighThreshold:
		return DriftBandHigh
	case score >= driftMediumThreshold:
		return DriftBandMedium
	default:
		return DriftBandLow
	}
}

func targetBetaCenter(target string) (float64, bool) {
	switch strings.ToUpper(strings.TrimSpace(target)) {
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

func topDrivers(comps map[string]float64, terms map[string]float64, k int) []Driver {
	type pair struct {
		code         string
		contribution float64
	}
	pairs := make([]pair, 0, len(comps))
	for code, c := range comps {
		if c <= 0 {
			continue
		}
		pairs = append(pairs, pair{code: code, contribution: c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].contribution == pairs[j].contribution {
			return pairs[i].code < pairs[j].code
		}
		return pairs[i].contribution > pairs[j].contribution
	})
	if k > len(pairs) {
		k = len(pairs)
	}
	out := make([]Driver, 0, k)
	for i := 0; i < k; i++ {
		code := pairs[i].code
		out = append(out, Driver{
			Code:         code,
			Label:        driverLabel(code),
			Contribution: round6(pairs[i].contribution),
			Value:        round6(terms[code]),
			Detail:       driverDetail(code),
		})
	}
	return out
}

func driverLabel(code string) string {
	switch code {
	case "BETA":
		return "Weighted beta"
	case "MISALIGNMENT":
		return "Distance from risk target"
	case "POSITION_CONCENTRATION":
		return "Position concentration"
	case "SECTOR_CONCENTRATION":
		return "Sector concentration"
	case "VOLATILITY":
		return "Recent volatility"
	case "ALLOCATION_IMBALANCE":
		return "Allocation imbalance"
	case "DRAWDOWN":
		return "Drawdown proxy"
	case "CORRELATION":
		return "Within-sector correlation proxy"
	case "SENTIMENT":
		return "News sentiment drag"
	default:
		return code
	}
}

func driverDetail(code string) string {
	switch code {
	case "BETA":
		return "How sensitive the bundle may be to broad market moves vs a beta near 1.0."
	case "MISALIGNMENT":
		return "Gap between estimated weighted beta and the reference for the stated target risk."
	case "POSITION_CONCENTRATION":
		return "Herfindahl-style index across positions; higher = fewer names dominate."
	case "SECTOR_CONCENTRATION":
		return "Herfindahl-style index across sectors when sector data is available."
	case "VOLATILITY":
		return "Weighted recent intra-day volatility proxy when supplied."
	case "ALLOCATION_IMBALANCE":
		return "Gini-style inequality across position weights (0 = perfectly even)."
	case "DRAWDOWN":
		return "Heuristic drawdown proxy combining beta and volatility."
	case "CORRELATION":
		return "Sector-overlap proxy: how much sector mass concentrates beyond the 1/n baseline."
	case "SENTIMENT":
		return "Weighted bearish drag from per-holding news sentiment when supplied."
	default:
		return ""
	}
}

func roundComps(in map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = round6(v)
	}
	return out
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// Package portfoliorisk implements a deterministic, explainable portfolio
// risk feature set + composite risk and drift scoring.
//
// The package intentionally avoids ML black-boxes. Each feature is a simple,
// well-known statistic (weighted beta, Herfindahl, Gini, sector overlap,
// volatility-weighted drawdown proxy, etc.) so the pipeline stays auditable
// and easy to demo / explain. Outputs are educational and not advisory.
package portfoliorisk

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Holding is one position used by feature extraction. Beta / Sector /
// Volatility24h / Sentiment are optional; missing fields lower confidence
// rather than poison the score with bogus defaults.
type Holding struct {
	Symbol        string  `json:"symbol"`
	Allocation    float64 `json:"allocation"`
	Beta          string  `json:"beta,omitempty"`
	Sector        string  `json:"sector,omitempty"`
	Volatility24h string  `json:"volatility_24h,omitempty"`
	// Sentiment is an optional bucket: BULLISH / BEARISH / NEUTRAL / MIXED.
	Sentiment string `json:"sentiment,omitempty"`
}

// Features is the bundle of normalised structural features for a holdings
// snapshot. All weights are normalised so that Σ wᵢ = 1.
type Features struct {
	NumPositions       int       `json:"num_positions"`
	NormalizedWeights  []float64 `json:"normalized_weights,omitempty"`
	Symbols            []string  `json:"symbols,omitempty"`
	WeightedBeta       float64   `json:"weighted_beta"`
	PositionHHI        float64   `json:"position_hhi"`
	Top3Concentration  float64   `json:"top3_concentration"`
	EffectiveN         float64   `json:"effective_n"`
	AllocationGini     float64   `json:"allocation_gini"`
	SectorHHI          float64   `json:"sector_hhi"`
	SectorCount        int       `json:"sector_count"`
	WeightedVolatility float64   `json:"weighted_volatility"`
	DrawdownProxy      float64   `json:"drawdown_proxy"`
	CorrelationProxy   float64   `json:"correlation_proxy"`
	SentimentScore     float64   `json:"sentiment_score"`
	// Coverage fractions in [0, 1] = portion of weight with that field present.
	BetaCoverage       float64 `json:"beta_coverage"`
	SectorCoverage     float64 `json:"sector_coverage"`
	VolatilityCoverage float64 `json:"volatility_coverage"`
	SentimentCoverage  float64 `json:"sentiment_coverage"`
}

// ExtractFeatures normalises weights and computes every structural feature.
// Returns an error only when the input is structurally invalid (empty list,
// negative allocation, non-positive sum). Missing optional fields are fine.
func ExtractFeatures(holdings []Holding) (Features, error) {
	if len(holdings) == 0 {
		return Features{}, fmt.Errorf("holdings required")
	}
	sumW := 0.0
	for _, h := range holdings {
		if h.Allocation < 0 {
			return Features{}, fmt.Errorf("negative allocation for %s", h.Symbol)
		}
		sumW += h.Allocation
	}
	if sumW <= 0 {
		return Features{}, fmt.Errorf("allocations must sum to a positive value")
	}
	inv := 1.0 / sumW

	n := len(holdings)
	weights := make([]float64, n)
	symbols := make([]string, n)

	var (
		wBeta, hhi               float64
		volNumer, volDenom       float64
		drawdownNumer, drawdownD float64
		sentNumer, sentDenom     float64
		betaCovW, secCovW        float64
		volCovW, sentCovW        float64
	)
	sectorW := make(map[string]float64)
	knownSectorTotal := 0.0

	for i, h := range holdings {
		w := h.Allocation * inv
		weights[i] = w
		symbols[i] = strings.ToUpper(strings.TrimSpace(h.Symbol))
		hhi += w * w

		beta, betaOK := parseFloatField(h.Beta)
		if !betaOK {
			beta = 1.0 // neutral default for the weighted-beta math, no coverage credit
		} else {
			betaCovW += w
		}
		wBeta += beta * w

		if sec := strings.TrimSpace(h.Sector); sec != "" {
			sectorW[strings.ToUpper(sec)] += w
			secCovW += w
			knownSectorTotal += w
		}

		if vol, ok := parseFloatField(h.Volatility24h); ok && vol >= 0 {
			volNumer += vol * w
			volDenom += w
			volCovW += w
			drawdownNumer += beta * vol * w
			drawdownD += w
		}

		if s := sentimentScalar(h.Sentiment); s >= 0 {
			sentNumer += s * w
			sentDenom += w
			sentCovW += w
		}
	}

	// Sector HHI is computed only over the *known* portion so that fully-
	// missing sector info gives 0 (signalling "unknown") instead of a
	// spurious 1.0 ("everything in one bucket called UNKNOWN").
	sectorHHI := 0.0
	if knownSectorTotal > 0 {
		for _, sw := range sectorW {
			share := sw / knownSectorTotal
			sectorHHI += share * share
		}
	}

	weightedVol := 0.0
	if volDenom > 0 {
		weightedVol = volNumer / volDenom
	}
	drawdown := 0.0
	if drawdownD > 0 {
		drawdown = drawdownNumer / drawdownD
	}
	sentiment := 0.0
	if sentDenom > 0 {
		sentiment = sentNumer / sentDenom
	}

	// Sector-overlap correlation proxy: how much sector mass concentrates
	// beyond the 1/n baseline that perfectly-spread holdings would have.
	correlation := 0.0
	if knownSectorTotal > 0 {
		correlation = math.Max(0, sectorHHI-1.0/float64(n))
	}

	feat := Features{
		NumPositions:       n,
		NormalizedWeights:  weights,
		Symbols:            symbols,
		WeightedBeta:       round6(wBeta),
		PositionHHI:        round6(hhi),
		Top3Concentration:  round6(topKConcentration(weights, 3)),
		EffectiveN:         round6(effectiveN(hhi)),
		AllocationGini:     round6(discreteGini(weights)),
		SectorHHI:          round6(sectorHHI),
		SectorCount:        len(sectorW),
		WeightedVolatility: round6(weightedVol),
		DrawdownProxy:      round6(drawdown),
		CorrelationProxy:   round6(correlation),
		SentimentScore:     round6(sentiment),
		BetaCoverage:       round6(betaCovW),
		SectorCoverage:     round6(secCovW),
		VolatilityCoverage: round6(volCovW),
		SentimentCoverage:  round6(sentCovW),
	}
	return feat, nil
}

// effectiveN = 1 / HHI; the equivalent number of equally-weighted holdings.
func effectiveN(hhi float64) float64 {
	if hhi <= 0 {
		return 0
	}
	return 1.0 / hhi
}

// topKConcentration returns the share of the K largest weights.
func topKConcentration(weights []float64, k int) float64 {
	if k <= 0 || len(weights) == 0 {
		return 0
	}
	cp := append([]float64(nil), weights...)
	sort.Sort(sort.Reverse(sort.Float64Slice(cp)))
	if k > len(cp) {
		k = len(cp)
	}
	sum := 0.0
	for i := 0; i < k; i++ {
		sum += cp[i]
	}
	return sum
}

// discreteGini computes the classic Gini coefficient on weights in [0, 1].
// 0 = perfectly even, →1 = single dominant weight.
func discreteGini(weights []float64) float64 {
	n := len(weights)
	if n < 2 {
		return 0
	}
	mean := 0.0
	for _, w := range weights {
		mean += w
	}
	mean /= float64(n)
	if mean == 0 {
		return 0
	}
	var sumAbs float64
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			d := weights[i] - weights[j]
			if d < 0 {
				d = -d
			}
			sumAbs += d
		}
	}
	g := sumAbs / (2.0 * float64(n) * float64(n) * mean)
	if g < 0 {
		g = 0
	}
	if g > 1 {
		g = 1
	}
	return g
}

// sentimentScalar maps a sentiment label to a [0, 1] "bearish drag" value.
// Higher = more bearish pressure on the portfolio.
// Returns -1 for unknown / empty so the caller can skip it cleanly.
func sentimentScalar(label string) float64 {
	switch strings.ToUpper(strings.TrimSpace(label)) {
	case "BULLISH":
		return 0.0
	case "NEUTRAL":
		return 0.5
	case "MIXED":
		return 0.6
	case "BEARISH":
		return 1.0
	default:
		return -1
	}
}

func parseFloatField(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func round6(x float64) float64 {
	return math.Round(x*1_000_000) / 1_000_000
}

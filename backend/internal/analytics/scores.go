package analytics

import (
	"math"
	"strings"
)

// DiversificationScore is 0–100 from normalized Shannon entropy of market-value weights.
// One holding or empty → 0; equal weights maximize toward 100.
func DiversificationScore(marketValues []float64) float64 {
	nPos := 0
	var total float64
	for _, v := range marketValues {
		if v > 0 {
			nPos++
			total += v
		}
	}
	if total <= 0 || nPos <= 1 {
		return 0
	}

	var h float64
	for _, v := range marketValues {
		if v <= 0 {
			continue
		}
		w := v / total
		h -= w * math.Log(w)
	}

	maxH := math.Log(float64(nPos))
	if maxH <= 0 {
		return 0
	}
	norm := h / maxH
	if norm < 0 {
		norm = 0
	}
	if norm > 1 {
		norm = 1
	}
	return math.Round(norm*10000) / 100
}

func assetVolProxy(assetType string) float64 {
	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "crypto", "cryptocurrency":
		return 0.55
	case "etf", "fund", "index":
		return 0.16
	case "bond", "fixed_income":
		return 0.06
	case "cash", "money_market":
		return 0.02
	default:
		return 0.22
	}
}

// VolatilityScore is 0–100 (higher = more volatile) using value-weighted
// annualized vol proxies by asset class. Reference vol 45% maps to 100.
func VolatilityScore(marketValues []float64, assetTypes []string) float64 {
	if len(marketValues) != len(assetTypes) {
		return 0
	}

	var total float64
	for _, v := range marketValues {
		if v > 0 {
			total += v
		}
	}
	if total <= 0 {
		return 0
	}

	var weighted float64
	for i, v := range marketValues {
		if v <= 0 {
			continue
		}
		w := v / total
		weighted += w * assetVolProxy(assetTypes[i])
	}

	const refVol = 0.45
	score := (weighted / refVol) * 100
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return math.Round(score*100) / 100
}

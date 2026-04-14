package prediction

import "wealthscope-ai/internal/ml"

// DriftHoldingsFromML maps legacy portfolio holdings into drift inputs (optional sector on DriftHolding stays empty).
func DriftHoldingsFromML(h []ml.PortfolioHolding) []DriftHolding {
	out := make([]DriftHolding, len(h))
	for i, x := range h {
		out[i] = DriftHolding{
			Symbol:     x.Symbol,
			Allocation: x.Allocation,
			Beta:       x.Beta,
		}
	}
	return out
}

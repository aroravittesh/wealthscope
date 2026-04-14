package portfoliosvc

import (
	"fmt"
	"sort"
	"strings"

	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/prediction"
)

// ChangesRequest is the JSON body for POST /portfolio/changes.
// Prior may be omitted or empty to indicate no snapshot for comparison.
type ChangesRequest struct {
	Current struct {
		Holdings []ml.PortfolioHolding `json:"holdings"`
	} `json:"current"`
	Prior struct {
		Holdings []ml.PortfolioHolding `json:"holdings"`
	} `json:"prior"`
}

// ChangesResponse describes deltas between two snapshots (neutral, informational).
type ChangesResponse struct {
	HasPriorSnapshot              bool     `json:"has_prior_snapshot"`
	ChangesSummary                string   `json:"changes_summary"`
	WeightedBetaDelta             float64  `json:"weighted_beta_delta"`
	PositionCountDelta            int      `json:"position_count_delta"`
	SymbolsAdded                  []string `json:"symbols_added"`
	SymbolsRemoved                []string `json:"symbols_removed"`
	PositionConcentrationHHIDelta float64  `json:"position_concentration_hhi_delta"`
	Notes                         []string `json:"notes"`
	Disclaimer                    string   `json:"disclaimer"`
}

// DescribeChanges compares current holdings to an optional prior snapshot.
func DescribeChanges(req ChangesRequest) (*ChangesResponse, error) {
	if len(req.Current.Holdings) == 0 {
		return nil, fmt.Errorf("current.holdings required")
	}

	resp := &ChangesResponse{
		Disclaimer: EducationalDisclaimer,
		Notes:      nil,
	}

	if len(req.Prior.Holdings) == 0 {
		resp.HasPriorSnapshot = false
		resp.ChangesSummary = "No prior portfolio snapshot was provided, so risk and concentration deltas cannot be computed. " +
			"Send prior.holdings from an earlier save to compare how the profile shifted."
		resp.Notes = []string{
			"Tip: persist the last holdings array client-side or from your backend to send as prior.holdings on the next call.",
		}
		return resp, nil
	}

	curDrift, err := prediction.PredictRiskDrift(prediction.DriftRequest{
		Holdings:   prediction.DriftHoldingsFromML(req.Current.Holdings),
		TargetRisk: "MEDIUM",
	})
	if err != nil {
		return nil, err
	}
	priorDrift, err := prediction.PredictRiskDrift(prediction.DriftRequest{
		Holdings:   prediction.DriftHoldingsFromML(req.Prior.Holdings),
		TargetRisk: "MEDIUM",
	})
	if err != nil {
		return nil, err
	}

	curSyms := symbolSet(req.Current.Holdings)
	priorSyms := symbolSet(req.Prior.Holdings)
	added, removed := diffSets(curSyms, priorSyms)

	wbDelta := round4(curDrift.WeightedBeta - priorDrift.WeightedBeta)
	hhiDelta := round4(curDrift.PositionHHI - priorDrift.PositionHHI)
	countDelta := len(req.Current.Holdings) - len(req.Prior.Holdings)

	resp.HasPriorSnapshot = true
	resp.WeightedBetaDelta = wbDelta
	resp.PositionCountDelta = countDelta
	resp.SymbolsAdded = added
	resp.SymbolsRemoved = removed
	resp.PositionConcentrationHHIDelta = hhiDelta

	resp.ChangesSummary = fmt.Sprintf(
		"Weighted beta moved by %+0.4f (current %.2f vs prior %.2f). "+
			"Position count changed by %+d. "+
			"Position concentration (HHI) changed by %+0.4f (higher often means fewer names carry more weight).",
		wbDelta, curDrift.WeightedBeta, priorDrift.WeightedBeta,
		countDelta,
		hhiDelta,
	)

	if len(added) > 0 {
		resp.Notes = append(resp.Notes, "Symbols present now that were not in the prior snapshot: "+strings.Join(added, ", ")+".")
	}
	if len(removed) > 0 {
		resp.Notes = append(resp.Notes, "Symbols that appeared in the prior snapshot but not in the current one: "+strings.Join(removed, ", ")+".")
	}
	if len(added) == 0 && len(removed) == 0 && countDelta == 0 {
		resp.Notes = append(resp.Notes, "Symbol set is unchanged; any shift comes from weight or beta inputs.")
	}

	return resp, nil
}

func symbolSet(h []ml.PortfolioHolding) map[string]struct{} {
	m := make(map[string]struct{})
	for _, x := range h {
		s := strings.ToUpper(strings.TrimSpace(x.Symbol))
		if s != "" {
			m[s] = struct{}{}
		}
	}
	return m
}

func diffSets(current, prior map[string]struct{}) (added, removed []string) {
	for s := range current {
		if _, ok := prior[s]; !ok {
			added = append(added, s)
		}
	}
	for s := range prior {
		if _, ok := current[s]; !ok {
			removed = append(removed, s)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

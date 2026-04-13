package portfolioexplain

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/prediction"
)

// Request matches POST /portfolio/explain JSON.
type Request struct {
	Holdings   []ml.PortfolioHolding `json:"holdings"`
	TargetRisk string                `json:"target_risk"`
}

// Response is frontend-friendly structured copy (educational, not advice).
type Response struct {
	Summary              string   `json:"summary"`
	TopRisks             []string `json:"top_risks"`
	ConcentrationWarning string   `json:"concentration_warning"`
	DiversificationNote  string   `json:"diversification_note"`
	RiskAlignment        string   `json:"risk_alignment"`
	NeutralGuidance      []string `json:"neutral_guidance"`
}

// Explain turns holdings + target into neutral educational text using risk + drift logic.
func Explain(req Request) (*Response, error) {
	if len(req.Holdings) == 0 {
		return nil, fmt.Errorf("holdings required")
	}

	driftReq := prediction.DriftRequest{
		TargetRisk: req.TargetRisk,
		Holdings:   toDriftHoldings(req.Holdings),
	}
	drift, err := prediction.PredictRiskDrift(driftReq)
	if err != nil {
		return nil, err
	}

	report := ml.ScorePortfolio(req.Holdings)
	n := len(req.Holdings)
	wb := drift.WeightedBeta
	hhi := drift.PositionHHI
	secHHI := drift.SectorHHI
	target := strings.ToUpper(strings.TrimSpace(req.TargetRisk))

	summary := fmt.Sprintf(
		"This mix has an estimated weighted beta of about %.2f: that is a rough gauge of how sensitive the bundle may be to broad market moves compared with a beta near 1.0. "+
			"There are %d position(s) in the snapshot. The implied volatility profile is in the %s band on a simple beta scale. "+
			"These labels are educational only and not a forecast.",
		wb, n, strings.ToLower(string(report.Level)),
	)

	topRisks := buildTopRisks(wb, hhi, secHHI, drift.DriftLevel, target)
	concWarn := concentrationCopy(hhi)
	divNote := diversificationCopy(n, hhi)
	align := alignmentCopy(drift.DriftLevel, target, wb)
	guidance := neutralGuidanceBullets()

	return &Response{
		Summary:              summary,
		TopRisks:             topRisks,
		ConcentrationWarning: concWarn,
		DiversificationNote:  divNote,
		RiskAlignment:        align,
		NeutralGuidance:      guidance,
	}, nil
}

func toDriftHoldings(h []ml.PortfolioHolding) []prediction.DriftHolding {
	out := make([]prediction.DriftHolding, len(h))
	for i, x := range h {
		out[i] = prediction.DriftHolding{
			Symbol:     x.Symbol,
			Allocation: x.Allocation,
			Beta:       x.Beta,
		}
	}
	return out
}

func buildTopRisks(wb, hhi, secHHI float64, driftLevel, target string) []string {
	risks := make([]string, 0, 4)
	if wb >= 1.4 {
		risks = append(risks, "Weighted beta is elevated, meaning day-to-day moves may be amplified versus the overall market on average.")
	} else if wb <= 0.55 {
		risks = append(risks, "Weighted beta is low relative to many equity mixes, so participation in strong up-markets may be muted in relative terms.")
	}

	if hhi >= 0.5 {
		risks = append(risks, "Position weights are concentrated: a large share sits in few names, so single-company outcomes can sway the whole snapshot.")
	} else if hhi >= 0.34 {
		risks = append(risks, "There is moderate concentration by weight; diversification is not evenly spread across many small slices.")
	}

	if secHHI >= 0.55 {
		risks = append(risks, "Sector weights are concentrated when sectors are known, which can cluster exposure to one part of the economy.")
	}

	switch driftLevel {
	case string(prediction.DriftHigh):
		risks = append(risks, fmt.Sprintf("Versus a stated %s risk target, this snapshot sits far from the simple reference profile used in the model.", strings.ToLower(target)))
	case string(prediction.DriftMedium):
		risks = append(risks, fmt.Sprintf("Versus a %s risk target, there is a partial gap on the metrics used here—not extreme, but not a tight match.", strings.ToLower(target)))
	}

	if len(risks) > 4 {
		risks = risks[:4]
	}
	if len(risks) == 0 {
		risks = append(risks, "On this coarse pass, no standout structural flags beyond ordinary equity variability appeared.")
	}
	return risks
}

func concentrationCopy(hhi float64) string {
	switch {
	case hhi >= 0.55:
		return "Concentration by weight is high (Herfindahl-style index near a single-name book). That raises sensitivity to individual holdings without implying any action."
	case hhi >= 0.34:
		return "Concentration is noticeable: a few positions carry most of the weight. Reviewing weight balance is a common educational step."
	default:
		return "Position-level concentration looks moderate or broad for this snapshot; still, concentration can change quickly if weights shift."
	}
}

func diversificationCopy(n int, hhi float64) string {
	if n >= 8 && hhi < 0.2 {
		return "Many small weights typically spread idiosyncratic risk more evenly, though overlap in factors or sectors can still cluster risk."
	}
	if n <= 3 {
		return "A small number of lines means diversification depends heavily on how different those names behave from each other and from the rest of the market."
	}
	return "Diversification is about both how many names you hold and how uncorrelated they are; count alone does not capture overlap in themes or sectors."
}

func alignmentCopy(driftLevel, target string, wb float64) string {
	t := strings.ToLower(target)
	switch driftLevel {
	case string(prediction.DriftLow):
		return fmt.Sprintf("Relative to a %s risk target, the simple drift check reads as close on the metrics used (weighted beta about %.2f).", t, wb)
	case string(prediction.DriftMedium):
		return fmt.Sprintf("Relative to a %s risk target, alignment is mixed: some metrics line up and others diverge (weighted beta about %.2f).", t, wb)
	case string(prediction.DriftHigh):
		return fmt.Sprintf("Relative to a %s risk target, this snapshot shows meaningful distance on the heuristic used—mainly beta and concentration signals (weighted beta about %.2f).", t, wb)
	default:
		return fmt.Sprintf("Target profile stated: %s. Use the figures above as context only.", t)
	}
}

func neutralGuidanceBullets() []string {
	return []string{
		"Many investors compare weighted beta and concentration when learning how a mix might behave in different market phases.",
		"Reading filings, index definitions, and fund fact sheets can clarify overlap that raw position counts miss.",
		"Educational tools like this one do not replace a personal risk budget or professional advice where appropriate.",
	}
}

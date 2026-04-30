package explain

import (
	"fmt"
	"strings"

	"wealthscope-ai/internal/portfoliorisk"
)

// RiskInputs is the bundle from a portfoliorisk.Assessment used to build the
// risk explanation envelope.
type RiskInputs struct {
	Level      string // LOW | MEDIUM | HIGH
	Score      float64
	Confidence float64
	Drivers    []portfoliorisk.Driver
}

// DriftInputs is the bundle from a portfoliorisk.Assessment used to build the
// drift explanation envelope.
type DriftInputs struct {
	Level         string // LOW_DRIFT | MEDIUM_DRIFT | HIGH_DRIFT
	Score         float64
	Confidence    float64
	Target        string  // LOW | MEDIUM | HIGH
	WeightedBeta  float64
	CenterBeta    float64
	Misalignment  float64
	Drivers       []portfoliorisk.Driver
}

// BuildRiskExplanation constructs a structured explanation for the composite
// portfolio risk score.
func BuildRiskExplanation(in RiskInputs) Explanation {
	reasons := []string{
		fmt.Sprintf("Composite risk score %.2f maps to %s on the calibrated scale.", in.Score, strings.ToUpper(in.Level)),
	}
	signals := make([]Signal, 0, len(in.Drivers))
	for _, d := range in.Drivers {
		reasons = append(reasons,
			fmt.Sprintf("%s contributes %.2f (feature value %.2f).", d.Label, d.Contribution, d.Value))
		signals = append(signals, driverToSignal("RISK_DRIVER_", d))
	}
	return Explanation{
		Code:       riskReasonCode(in.Level, in.Drivers),
		Summary:    riskSummary(in.Level, in.Drivers),
		Confidence: in.Confidence,
		Source:     "portfoliorisk_composite",
		Reasons:    reasons,
		TopSignals: signals,
		Disclaimer: EducationalDisclaimer,
	}
}

// BuildDriftExplanation constructs a structured explanation for the
// risk-target drift score.
func BuildDriftExplanation(in DriftInputs) Explanation {
	level := strings.ToUpper(in.Level)
	target := strings.ToUpper(strings.TrimSpace(in.Target))
	reasons := []string{
		fmt.Sprintf("Drift score %.2f maps to %s versus stated target %s.", in.Score, level, target),
	}
	if in.CenterBeta > 0 {
		reasons = append(reasons,
			fmt.Sprintf("Weighted beta %.2f vs target reference %.2f (gap term %.2f).",
				in.WeightedBeta, in.CenterBeta, in.Misalignment))
	}
	signals := make([]Signal, 0, len(in.Drivers))
	for _, d := range in.Drivers {
		reasons = append(reasons, fmt.Sprintf("%s contributes %.2f.", d.Label, d.Contribution))
		signals = append(signals, driverToSignal("DRIFT_DRIVER_", d))
	}
	return Explanation{
		Code:       driftReasonCode(level, in.Drivers),
		Summary:    driftSummary(level, target, in.WeightedBeta, in.CenterBeta, in.Drivers),
		Confidence: in.Confidence,
		Source:     "portfoliorisk_drift",
		Reasons:    reasons,
		TopSignals: signals,
		Disclaimer: EducationalDisclaimer,
	}
}

func driverToSignal(prefix string, d portfoliorisk.Driver) Signal {
	return Signal{
		Code:   prefix + d.Code,
		Label:  d.Label,
		Score:  d.Contribution,
		Detail: d.Detail,
	}
}

func riskReasonCode(level string, drivers []portfoliorisk.Driver) string {
	level = strings.ToUpper(level)
	if len(drivers) == 0 {
		return fmt.Sprintf("RISK_%s_GENERIC", level)
	}
	return fmt.Sprintf("RISK_%s_DRIVEN_BY_%s", level, drivers[0].Code)
}

func driftReasonCode(level string, drivers []portfoliorisk.Driver) string {
	level = strings.ToUpper(level)
	if len(drivers) == 0 {
		return fmt.Sprintf("%s_GENERIC", level)
	}
	return fmt.Sprintf("%s_DRIVEN_BY_%s", level, drivers[0].Code)
}

func riskSummary(level string, drivers []portfoliorisk.Driver) string {
	low := strings.ToLower(level)
	if len(drivers) == 0 {
		return fmt.Sprintf("Portfolio risk reads as %s on the composite scale.", low)
	}
	names := topDriverLabels(drivers, 2)
	return fmt.Sprintf("Portfolio risk reads as %s, mainly because of %s.", low, strings.Join(names, " and "))
}

func driftSummary(level, target string, wBeta, center float64, drivers []portfoliorisk.Driver) string {
	low := strings.ToLower(level)
	base := fmt.Sprintf("Drift vs %s target reads as %s (weighted beta %.2f vs %.2f reference)",
		target, low, wBeta, center)
	if len(drivers) == 0 {
		return base + "."
	}
	names := topDriverLabels(drivers, 2)
	return base + ", mainly because of " + strings.Join(names, " and ") + "."
}

func topDriverLabels(drivers []portfoliorisk.Driver, k int) []string {
	if k > len(drivers) {
		k = len(drivers)
	}
	out := make([]string, 0, k)
	for i := 0; i < k; i++ {
		out = append(out, strings.ToLower(drivers[i].Label))
	}
	return out
}

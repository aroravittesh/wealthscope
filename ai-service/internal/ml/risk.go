package ml

import (
    "fmt"
    "strconv"
)

type RiskLevel string

const (
    RiskLow    RiskLevel = "LOW"
    RiskMedium RiskLevel = "MEDIUM"
    RiskHigh   RiskLevel = "HIGH"
)

type PortfolioHolding struct {
    Symbol     string  `json:"symbol"`
    Allocation float64 `json:"allocation"`
    Beta       string  `json:"beta"`
}

type RiskReport struct {
    Score       float64   `json:"Score"`
    Level       RiskLevel `json:"Level"`
    Explanation string    `json:"Explanation"`
}

func ScorePortfolio(holdings []PortfolioHolding) RiskReport {
    weightedBeta := 0.0

    for _, h := range holdings {
        beta, err := strconv.ParseFloat(h.Beta, 64)
        if err != nil {
            beta = 1.0
        }
        weightedBeta += beta * h.Allocation
    }

    var level RiskLevel
    var explanation string

    switch {
    case weightedBeta >= 1.5:
        level = RiskHigh
        explanation = fmt.Sprintf(
            "Your portfolio has a weighted beta of %.2f, indicating high volatility relative to the market. Consider adding stable, low-beta assets.",
            weightedBeta,
        )
    case weightedBeta >= 0.8:
        level = RiskMedium
        explanation = fmt.Sprintf(
            "Your portfolio has a weighted beta of %.2f, broadly in line with market movements. A balanced profile.",
            weightedBeta,
        )
    default:
        level = RiskLow
        explanation = fmt.Sprintf(
            "Your portfolio has a weighted beta of %.2f, suggesting low volatility. May underperform in strong bull markets.",
            weightedBeta,
        )
    }

    return RiskReport{
        Score:       weightedBeta,
        Level:       level,
        Explanation: explanation,
    }
}
package explain

import (
	"fmt"
	"math"
	"strings"

	"wealthscope-ai/internal/finsentiment"
)

// SentimentInputs is the bundle of facts the news-sentiment aggregator already
// computes, repackaged for explanation construction.
type SentimentInputs struct {
	Symbol       string
	Overall      string // BULLISH | BEARISH | NEUTRAL | MIXED
	Mean         float64
	Dispersion   float64
	Confidence   float64
	ArticleCount int
	TopPositive  string
	TopNegative  string
	DrivingTerms []finsentiment.TermHit
}

// BuildSentimentExplanation builds a finance-aware sentiment explanation
// envelope from the inputs the aggregator already has on hand.
func BuildSentimentExplanation(in SentimentInputs) Explanation {
	posSignals, negSignals := splitTermSignals(in.DrivingTerms)

	reasons := []string{
		fmt.Sprintf("Aggregated %d recent headline(s) for %s.", in.ArticleCount, in.Symbol),
		fmt.Sprintf("Mean polarity %+.2f, dispersion %.2f drove the %s classification.",
			in.Mean, in.Dispersion, strings.ToUpper(in.Overall)),
	}
	if in.TopPositive != "" {
		reasons = append(reasons, fmt.Sprintf("Top positive headline: %q", in.TopPositive))
	}
	if in.TopNegative != "" {
		reasons = append(reasons, fmt.Sprintf("Top negative headline: %q", in.TopNegative))
	}

	exp := Explanation{
		Code:       sentimentReasonCode(in.Overall),
		Summary:    buildSentimentSummary(in),
		Confidence: in.Confidence,
		Source:     "finance_aware_lexicon",
		Reasons:    reasons,
		Disclaimer: EducationalDisclaimer,
	}
	exp.TopSignals = append(exp.TopSignals, posSignals...)
	exp.TopSignals = append(exp.TopSignals, negSignals...)
	return exp
}

// SignalsFromTerms exposes the same term-to-signal mapping for callers that
// want to attach top_positive_signals / top_negative_signals directly without
// the full Explanation envelope.
func SignalsFromTerms(terms []finsentiment.TermHit) (positive, negative []Signal) {
	return splitTermSignals(terms)
}

func splitTermSignals(terms []finsentiment.TermHit) (positive, negative []Signal) {
	for _, t := range terms {
		sig := Signal{
			Code:     sentimentSignalCode(t),
			Label:    t.Term,
			Score:    math.Abs(t.Polarity),
			Polarity: t.Polarity,
			Detail:   sentimentSignalDetail(t),
		}
		if t.Polarity >= 0 {
			positive = append(positive, sig)
		} else {
			negative = append(negative, sig)
		}
	}
	return positive, negative
}

func sentimentSignalCode(t finsentiment.TermHit) string {
	if strings.Contains(t.Term, " ") {
		if t.Polarity >= 0 {
			return "FIN_PHRASE_BULLISH"
		}
		return "FIN_PHRASE_BEARISH"
	}
	if t.Polarity >= 0 {
		return "FIN_TERM_BULLISH"
	}
	return "FIN_TERM_BEARISH"
}

func sentimentSignalDetail(t finsentiment.TermHit) string {
	direction := "bullish"
	if t.Polarity < 0 {
		direction = "bearish"
	}
	return fmt.Sprintf("%q contributed a %s polarity of %+.2f.", t.Term, direction, t.Polarity)
}

func sentimentReasonCode(overall string) string {
	switch strings.ToUpper(overall) {
	case "BULLISH":
		return "SENT_BULLISH_DOMINANT_POSITIVE"
	case "BEARISH":
		return "SENT_BEARISH_DOMINANT_NEGATIVE"
	case "MIXED":
		return "SENT_MIXED_HIGH_DISPERSION"
	case "NEUTRAL":
		return "SENT_NEUTRAL_NEAR_ZERO"
	default:
		return "SENT_UNKNOWN"
	}
}

func buildSentimentSummary(in SentimentInputs) string {
	switch strings.ToUpper(in.Overall) {
	case "BULLISH":
		return fmt.Sprintf("%s news skews bullish: positive headlines outweighed negatives across %d article(s) (mean polarity %+.2f).",
			in.Symbol, in.ArticleCount, in.Mean)
	case "BEARISH":
		return fmt.Sprintf("%s news skews bearish: negative headlines outweighed positives across %d article(s) (mean polarity %+.2f).",
			in.Symbol, in.ArticleCount, in.Mean)
	case "MIXED":
		return fmt.Sprintf("%s news is mixed: positive and negative coverage are present in similar weight (dispersion %.2f across %d article(s)).",
			in.Symbol, in.Dispersion, in.ArticleCount)
	case "NEUTRAL":
		return fmt.Sprintf("%s news reads as neutral: no decisive positive or negative tilt across %d article(s).",
			in.Symbol, in.ArticleCount)
	default:
		return fmt.Sprintf("%s sentiment classification: %s across %d article(s).", in.Symbol, in.Overall, in.ArticleCount)
	}
}

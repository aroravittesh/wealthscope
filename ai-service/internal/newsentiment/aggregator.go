package newsentiment

import (
	"fmt"
	"math"
	"strings"

	"wealthscope-ai/internal/explain"
	"wealthscope-ai/internal/finsentiment"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
)

// Response is the JSON shape for GET /news-sentiment/:symbol.
type Response struct {
	Symbol             string  `json:"symbol"`
	OverallSentiment   string  `json:"overall_sentiment"`
	Confidence         float64 `json:"confidence"`
	ArticleCount       int     `json:"article_count"`
	TopPositiveArticle string  `json:"top_positive_article"`
	TopNegativeArticle string  `json:"top_negative_article"`
	Summary            string  `json:"summary"`

	// Explainability envelope (additive; safe for older clients to ignore).
	TopPositiveSignals []explain.Signal     `json:"top_positive_signals,omitempty"`
	TopNegativeSignals []explain.Signal     `json:"top_negative_signals,omitempty"`
	ReasonCode         string               `json:"reason_code,omitempty"`
	ReasoningSummary   string               `json:"reasoning_summary,omitempty"`
	ExplanationDetail  *explain.Explanation `json:"explanation_detail,omitempty"`
}

// NewsFetcher loads articles (production uses market.GetMarketNews).
type NewsFetcher interface {
	FetchNews(symbol string) ([]market.NewsItem, error)
}

// LiveFetcher delegates to the existing market client.
type LiveFetcher struct{}

func (LiveFetcher) FetchNews(symbol string) ([]market.NewsItem, error) {
	return market.GetMarketNews(symbol)
}

const (
	// neutralBand defines the |mean polarity| threshold below which the
	// aggregate is NEUTRAL (or MIXED if there's also strong dispersion).
	neutralBand = 0.10
	// mixedDispersion is the stddev threshold above which split articles are MIXED.
	mixedDispersion = 0.50
	// articleSentBand is the per-article polarity threshold used to count
	// positive vs negative articles for MIXED detection.
	articleSentBand = 0.05
)

type articleScore struct {
	item     market.NewsItem
	score    finsentiment.Score
	polarity float64
	title    string
}

// Aggregate scores each article with the finance-aware scorer (title weighted
// more than description) and rolls polarities up to a ticker-level signal.
//
// Output buckets: BULLISH | BEARISH | NEUTRAL | MIXED. MIXED is emitted only
// when articles disagree (positive AND negative present) AND the mean is near
// zero AND dispersion is high.
func Aggregate(symbol string, articles []market.NewsItem) Response {
	sym := strings.TrimSpace(strings.ToUpper(symbol))
	if len(articles) == 0 {
		emptyExp := explain.Explanation{
			Code:       "SENT_NO_ARTICLES",
			Summary:    fmt.Sprintf("No recent articles were returned for %s, so no sentiment could be aggregated.", sym),
			Source:     "finance_aware_lexicon",
			Reasons:    []string{"News fetcher returned zero articles."},
			Disclaimer: explain.EducationalDisclaimer,
		}
		return Response{
			Symbol:           sym,
			OverallSentiment: string(ml.SentimentNeutral),
			Confidence:       0,
			ArticleCount:     0,
			Summary: "No recent articles were returned for this symbol, so no sentiment aggregation was possible. " +
				"This is informational, not investment advice.",
			ReasonCode:        emptyExp.Code,
			ReasoningSummary:  emptyExp.Summary,
			ExplanationDetail: &emptyExp,
		}
	}

	scored := make([]articleScore, 0, len(articles))
	allTerms := make([]finsentiment.TermHit, 0)
	var sumPol, sumSqPol float64
	var posCount, negCount int

	for _, a := range articles {
		s := finsentiment.ScoreArticle(a)
		scored = append(scored, articleScore{
			item:     a,
			score:    s,
			polarity: s.Polarity,
			title:    strings.TrimSpace(a.Title),
		})
		allTerms = append(allTerms, s.Terms...)
		sumPol += s.Polarity
		sumSqPol += s.Polarity * s.Polarity
		switch {
		case s.Polarity >= articleSentBand:
			posCount++
		case s.Polarity <= -articleSentBand:
			negCount++
		}
	}

	n := float64(len(scored))
	mean := sumPol / n
	variance := sumSqPol/n - mean*mean
	if variance < 0 {
		variance = 0
	}
	dispersion := math.Sqrt(variance)
	agreement := 1 - math.Min(1, dispersion)

	overall := classify(mean, dispersion, posCount, negCount)
	confidence := scoreConfidence(overall, mean, agreement, n)

	posIdx, negIdx := pickExtremes(scored)
	topPos, topNeg := "", ""
	if posIdx >= 0 && scored[posIdx].polarity >= articleSentBand {
		topPos = scored[posIdx].title
	}
	if negIdx >= 0 && scored[negIdx].polarity <= -articleSentBand {
		topNeg = scored[negIdx].title
	}
	if topPos != "" && topPos == topNeg {
		topNeg = ""
	}

	driving := mergeTopTerms(allTerms, 4)
	summary := buildSummary(sym, overall, mean, dispersion, len(scored), topPos, topNeg, driving)

	posSignals, negSignals := explain.SignalsFromTerms(driving)
	exp := explain.BuildSentimentExplanation(explain.SentimentInputs{
		Symbol:       sym,
		Overall:      string(overall),
		Mean:         mean,
		Dispersion:   dispersion,
		Confidence:   round2(confidence),
		ArticleCount: len(scored),
		TopPositive:  topPos,
		TopNegative:  topNeg,
		DrivingTerms: driving,
	})

	return Response{
		Symbol:             sym,
		OverallSentiment:   string(overall),
		Confidence:         round2(confidence),
		ArticleCount:       len(scored),
		TopPositiveArticle: topPos,
		TopNegativeArticle: topNeg,
		Summary:            summary,
		TopPositiveSignals: posSignals,
		TopNegativeSignals: negSignals,
		ReasonCode:         exp.Code,
		ReasoningSummary:   exp.Summary,
		ExplanationDetail:  &exp,
	}
}

func classify(mean, dispersion float64, posCount, negCount int) ml.Sentiment {
	if posCount > 0 && negCount > 0 && math.Abs(mean) < neutralBand && dispersion >= mixedDispersion {
		return ml.SentimentMixed
	}
	switch {
	case mean > neutralBand:
		return ml.SentimentBullish
	case mean < -neutralBand:
		return ml.SentimentBearish
	default:
		return ml.SentimentNeutral
	}
}

func scoreConfidence(overall ml.Sentiment, mean, agreement, n float64) float64 {
	c := 0.30 + 0.40*math.Abs(mean) + 0.20*agreement + 0.10*math.Min(1, n/5)
	if overall == ml.SentimentMixed {
		// MIXED is by definition a low-conviction bucket.
		c = math.Min(c, 0.55)
	}
	if overall == ml.SentimentNeutral && math.Abs(mean) <= neutralBand {
		c = math.Min(c, 0.55)
	}
	return clamp01(c)
}

// pickExtremes returns indices of the most positive and most negative articles.
// Both may be the same when only one article exists or all share polarity.
func pickExtremes(scored []articleScore) (posIdx, negIdx int) {
	if len(scored) == 0 {
		return -1, -1
	}
	posIdx, negIdx = 0, 0
	maxP, minP := scored[0].polarity, scored[0].polarity
	for i := 1; i < len(scored); i++ {
		p := scored[i].polarity
		if p > maxP {
			maxP, posIdx = p, i
		}
		if p < minP {
			minP, negIdx = p, i
		}
	}
	return posIdx, negIdx
}

func mergeTopTerms(all []finsentiment.TermHit, k int) []finsentiment.TermHit {
	if len(all) == 0 || k <= 0 {
		return nil
	}
	type acc struct {
		hit finsentiment.TermHit
		ord int
	}
	idx := map[string]int{}
	out := make([]acc, 0, len(all))
	for i, h := range all {
		if pos, ok := idx[h.Term]; ok {
			out[pos].hit.Polarity += h.Polarity
			continue
		}
		idx[h.Term] = len(out)
		out = append(out, acc{hit: h, ord: i})
	}
	// rank by abs polarity, stable on first-seen order
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			ai, aj := math.Abs(out[i].hit.Polarity), math.Abs(out[j].hit.Polarity)
			if aj > ai || (aj == ai && out[j].ord < out[i].ord) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if k > len(out) {
		k = len(out)
	}
	res := make([]finsentiment.TermHit, 0, k)
	for _, a := range out[:k] {
		res = append(res, a.hit)
	}
	return res
}

func buildSummary(
	sym string,
	overall ml.Sentiment,
	mean, dispersion float64,
	count int,
	topPos, topNeg string,
	driving []finsentiment.TermHit,
) string {
	dir := strings.ToLower(string(overall))
	var b strings.Builder
	fmt.Fprintf(&b,
		"Across %d recent headline(s) for %s, finance-aware sentiment tilts %s "+
			"(mean polarity %+.2f, dispersion %.2f). ",
		count, sym, dir, mean, dispersion,
	)
	if topPos != "" {
		b.WriteString("Strongest positive-leaning headline: \"")
		b.WriteString(truncate(topPos, 160))
		b.WriteString("\". ")
	}
	if topNeg != "" {
		b.WriteString("Strongest negative-leaning headline: \"")
		b.WriteString(truncate(topNeg, 160))
		b.WriteString("\". ")
	}
	if len(driving) > 0 {
		terms := make([]string, 0, len(driving))
		for _, t := range driving {
			sign := "+"
			if t.Polarity < 0 {
				sign = "-"
			}
			terms = append(terms, fmt.Sprintf("%s%s", sign, t.Term))
		}
		b.WriteString("Driving terms: ")
		b.WriteString(strings.Join(terms, ", "))
		b.WriteString(". ")
	}
	b.WriteString("This is an informational signal from a finance-aware lexical model, not a forecast and not investment advice.")
	return strings.TrimSpace(b.String())
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func round2(x float64) float64 {
	return math.Round(x*100) / 100
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

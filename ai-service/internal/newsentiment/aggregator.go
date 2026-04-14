package newsentiment

import (
	"fmt"
	"math"
	"strings"

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

type articleScore struct {
	item   market.NewsItem
	margin float64
	title  string
}

// Aggregate scores headline+description per article and rolls up to ticker level.
func Aggregate(symbol string, articles []market.NewsItem) Response {
	sym := strings.TrimSpace(strings.ToUpper(symbol))
	if len(articles) == 0 {
		return Response{
			Symbol:           sym,
			OverallSentiment: string(ml.SentimentNeutral),
			Confidence:       0,
			ArticleCount:     0,
			Summary: "No recent articles were returned for this symbol, so no sentiment aggregation was possible. " +
				"This is not investment advice.",
		}
	}

	scored := make([]articleScore, 0, len(articles))
	var sumMargin float64
	for _, a := range articles {
		text := strings.TrimSpace(a.Title + " " + a.Description)
		bull, bear := ml.LexicalSentimentScores(text)
		margin := float64(bull - bear)
		scored = append(scored, articleScore{item: a, margin: margin, title: strings.TrimSpace(a.Title)})
		sumMargin += margin
	}

	n := float64(len(scored))
	mean := sumMargin / n

	// Classify aggregate: use mean margin with a small dead band for NEUTRAL.
	const band = 0.25
	var overall ml.Sentiment
	switch {
	case mean > band:
		overall = ml.SentimentBullish
	case mean < -band:
		overall = ml.SentimentBearish
	default:
		overall = ml.SentimentNeutral
	}

	// Confidence: separation of mean (normalized) + sample size, capped.
	sep := math.Min(1.0, math.Abs(mean)/3.0)
	size := math.Min(1.0, n/5.0)
	confidence := math.Min(1.0, 0.35+0.45*sep+0.20*size)
	if overall == ml.SentimentNeutral && math.Abs(mean) <= band {
		confidence = math.Min(confidence, 0.55)
	}

	posIdx, negIdx := 0, 0
	maxM := scored[0].margin
	minM := scored[0].margin
	for i := 1; i < len(scored); i++ {
		if scored[i].margin > maxM {
			maxM = scored[i].margin
			posIdx = i
		}
		if scored[i].margin < minM {
			minM = scored[i].margin
			negIdx = i
		}
	}

	topPos := scored[posIdx].title
	topNeg := scored[negIdx].title
	if posIdx == negIdx {
		topNeg = ""
	}

	summary := buildSummary(sym, overall, mean, len(scored), topPos, topNeg)

	return Response{
		Symbol:             sym,
		OverallSentiment:   string(overall),
		Confidence:         round2(confidence),
		ArticleCount:       len(scored),
		TopPositiveArticle: topPos,
		TopNegativeArticle: topNeg,
		Summary:            summary,
	}
}

func buildSummary(sym string, overall ml.Sentiment, mean float64, count int, topPos, topNeg string) string {
	dir := strings.ToLower(string(overall))
	msg := fmt.Sprintf(
		"Across %d recent headline(s) for %s, the lexical baseline tilts %s (average bullish-minus-bearish hit margin %.2f). ",
		count, sym, dir, mean,
	)
	if topPos != "" {
		msg += "The strongest positive-scoring headline by this lexicon: \"" + truncate(topPos, 160) + "\". "
	}
	if topNeg != "" && topNeg != topPos {
		msg += "The strongest negative-scoring headline: \"" + truncate(topNeg, 160) + "\". "
	}
	msg += "This is a simple keyword-based snapshot, not a forecast and not investment advice."
	return strings.TrimSpace(msg)
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

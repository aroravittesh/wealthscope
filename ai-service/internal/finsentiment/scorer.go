package finsentiment

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"wealthscope-ai/internal/market"
)

// Bucket is the polarity classification of a single piece of text.
type Bucket string

const (
	Bullish Bucket = "BULLISH"
	Bearish Bucket = "BEARISH"
	Neutral Bucket = "NEUTRAL"
)

// TermHit is one matched lexicon term and its signed contribution.
// Negated unigrams appear with a "!" prefix and reduced magnitude.
type TermHit struct {
	Term     string
	Polarity float64
}

// Score is the per-text result from ScoreText / ScoreArticle.
//
//   - Polarity: signed score in [-1.0, +1.0]
//   - Bullish/Bearish: weighted unsigned magnitudes (after negation flips)
//   - Terms: matched terms in original encounter order
type Score struct {
	Polarity float64
	Bullish  float64
	Bearish  float64
	Terms    []TermHit
}

const (
	// polaritySaturation maps a weighted (bull-bear) margin of this magnitude to ±1.
	polaritySaturation = 4.0
	// titleWeight makes article titles more influential than descriptions.
	titleWeight = 1.5
	// neutralBand: |polarity| at or below this is treated as Neutral by Bucket().
	neutralBand = 0.10
)

var tokenizer = regexp.MustCompile(`[a-z0-9'\-]+`)

// init sorts phrase lists longest-first so multi-word matches are preferred.
func init() {
	sortPhrasesLongestFirst(bullishPhrases)
	sortPhrasesLongestFirst(bearishPhrases)
}

func sortPhrasesLongestFirst(ps []Phrase) {
	sort.SliceStable(ps, func(i, j int) bool {
		return len(ps[i].Tokens) > len(ps[j].Tokens)
	})
}

// ScoreText returns the polarity score of an arbitrary text using the finance lexicon.
// Empty / non-alpha text returns the zero Score (Polarity 0.0).
func ScoreText(text string) Score {
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return Score{}
	}
	consumed := make([]bool, len(tokens))

	var bull, bear float64
	var hits []TermHit

	scanPhrases(tokens, consumed, bullishPhrases, &bull, &bear, &hits)
	scanPhrases(tokens, consumed, bearishPhrases, &bull, &bear, &hits)

	for i, tok := range tokens {
		if consumed[i] {
			continue
		}
		if w, ok := bullishUnigrams[tok]; ok {
			if isNegated(tokens, i, consumed) {
				bear += w * negationDamp
				hits = append(hits, TermHit{Term: "!" + tok, Polarity: -w * negationDamp})
			} else {
				bull += w
				hits = append(hits, TermHit{Term: tok, Polarity: w})
			}
			continue
		}
		if w, ok := bearishUnigrams[tok]; ok {
			if isNegated(tokens, i, consumed) {
				bull += w * negationDamp
				hits = append(hits, TermHit{Term: "!" + tok, Polarity: w * negationDamp})
			} else {
				bear += w
				hits = append(hits, TermHit{Term: tok, Polarity: -w})
			}
		}
	}

	polarity := clamp((bull-bear)/polaritySaturation, -1, 1)
	return Score{Polarity: polarity, Bullish: bull, Bearish: bear, Terms: hits}
}

// ScoreArticle scores a NewsItem with the title weighted more heavily than the description.
func ScoreArticle(item market.NewsItem) Score {
	titleScore := ScoreText(item.Title)
	descScore := ScoreText(item.Description)

	bull := titleScore.Bullish*titleWeight + descScore.Bullish
	bear := titleScore.Bearish*titleWeight + descScore.Bearish

	merged := make([]TermHit, 0, len(titleScore.Terms)+len(descScore.Terms))
	merged = append(merged, titleScore.Terms...)
	merged = append(merged, descScore.Terms...)

	polarity := clamp((bull-bear)/polaritySaturation, -1, 1)
	return Score{Polarity: polarity, Bullish: bull, Bearish: bear, Terms: merged}
}

// Bucket classifies a polarity into Bullish/Bearish/Neutral using the neutral band.
func (s Score) Bucket() Bucket {
	switch {
	case s.Polarity > neutralBand:
		return Bullish
	case s.Polarity < -neutralBand:
		return Bearish
	default:
		return Neutral
	}
}

// TopTerms returns up to k unique signed terms ranked by absolute polarity.
// Ties are broken by first occurrence.
func (s Score) TopTerms(k int) []TermHit {
	if k <= 0 || len(s.Terms) == 0 {
		return nil
	}
	type indexed struct {
		hit TermHit
		ord int
	}
	seen := make(map[string]int) // term -> index in agg
	agg := make([]indexed, 0, len(s.Terms))
	for i, h := range s.Terms {
		if idx, ok := seen[h.Term]; ok {
			agg[idx].hit.Polarity += h.Polarity
			continue
		}
		seen[h.Term] = len(agg)
		agg = append(agg, indexed{hit: h, ord: i})
	}
	sort.SliceStable(agg, func(i, j int) bool {
		ai, aj := math.Abs(agg[i].hit.Polarity), math.Abs(agg[j].hit.Polarity)
		if ai != aj {
			return ai > aj
		}
		return agg[i].ord < agg[j].ord
	})
	if k > len(agg) {
		k = len(agg)
	}
	out := make([]TermHit, 0, k)
	for _, a := range agg[:k] {
		out = append(out, a.hit)
	}
	return out
}

func tokenize(text string) []string {
	if text == "" {
		return nil
	}
	return tokenizer.FindAllString(strings.ToLower(text), -1)
}

func scanPhrases(tokens []string, consumed []bool, phrases []Phrase, bull, bear *float64, hits *[]TermHit) {
	for _, p := range phrases {
		if len(p.Tokens) == 0 || len(p.Tokens) > len(tokens) {
			continue
		}
		for i := 0; i+len(p.Tokens) <= len(tokens); i++ {
			if anyConsumed(consumed, i, len(p.Tokens)) {
				continue
			}
			if matchAt(tokens, i, p.Tokens) {
				markConsumed(consumed, i, len(p.Tokens))
				if p.Weight >= 0 {
					*bull += p.Weight
				} else {
					*bear += -p.Weight
				}
				*hits = append(*hits, TermHit{
					Term:     strings.Join(p.Tokens, " "),
					Polarity: p.Weight,
				})
			}
		}
	}
}

func matchAt(tokens []string, start int, phrase []string) bool {
	for k, want := range phrase {
		if tokens[start+k] != want {
			return false
		}
	}
	return true
}

func anyConsumed(consumed []bool, start, n int) bool {
	for k := 0; k < n; k++ {
		if consumed[start+k] {
			return true
		}
	}
	return false
}

func markConsumed(consumed []bool, start, n int) {
	for k := 0; k < n; k++ {
		consumed[start+k] = true
	}
}

// isNegated reports whether tokens[i] is preceded by a negator inside the negation window.
// Consumed positions (already-matched phrases) do not block the lookback.
func isNegated(tokens []string, i int, _ []bool) bool {
	start := i - negationWindow
	if start < 0 {
		start = 0
	}
	for k := start; k < i; k++ {
		if _, ok := negators[tokens[k]]; ok {
			return true
		}
	}
	return false
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

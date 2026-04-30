// Package finsentiment provides a lightweight, finance-aware sentiment scorer.
//
// Design goals:
//   - Deterministic, no external services or model files.
//   - Phrase-aware (multi-word finance terms beat unigram overlap).
//   - Weighted (strong signals like "tumble" outweigh moderate ones like "down").
//   - Negation-aware over a small left-window.
//
// The lexicon is intentionally kept curated and small; it is far easier to
// reason about and to swap for a learned classifier later than a sprawling list.
package finsentiment

// Phrase is a multi-token finance phrase with a signed weight.
// Positive weight = bullish, negative = bearish.
type Phrase struct {
	Tokens []string
	Weight float64
}

// strong/moderate weights for unigrams. Phrases use their own explicit weight.
const (
	weakSignal     = 1.0
	strongSignal   = 2.0
	negationWindow = 3
	negationDamp   = 0.5 // half-weight when negated
)

// bullishPhrases — checked before unigrams; longest-first ordering matters.
// Weights chosen so a single decisive phrase can drive an article's polarity.
var bullishPhrases = []Phrase{
	{Tokens: []string{"raised", "full", "year", "guidance"}, Weight: 3.0},
	{Tokens: []string{"raises", "full", "year", "guidance"}, Weight: 3.0},
	{Tokens: []string{"filed", "for", "ipo"}, Weight: 1.5},
	{Tokens: []string{"all", "time", "high"}, Weight: 2.5},
	{Tokens: []string{"all-time", "high"}, Weight: 2.5},
	{Tokens: []string{"record", "high"}, Weight: 2.0},
	{Tokens: []string{"record", "highs"}, Weight: 2.0},
	{Tokens: []string{"raised", "guidance"}, Weight: 2.5},
	{Tokens: []string{"raises", "guidance"}, Weight: 2.5},
	{Tokens: []string{"raised", "outlook"}, Weight: 2.0},
	{Tokens: []string{"raises", "outlook"}, Weight: 2.0},
	{Tokens: []string{"boosted", "outlook"}, Weight: 2.0},
	{Tokens: []string{"beat", "estimates"}, Weight: 2.5},
	{Tokens: []string{"beats", "estimates"}, Weight: 2.5},
	{Tokens: []string{"beat", "expectations"}, Weight: 2.0},
	{Tokens: []string{"beats", "expectations"}, Weight: 2.0},
	{Tokens: []string{"exceeded", "estimates"}, Weight: 2.5},
	{Tokens: []string{"exceeds", "estimates"}, Weight: 2.5},
	{Tokens: []string{"exceed", "estimates"}, Weight: 2.0},
	{Tokens: []string{"blew", "past"}, Weight: 2.0},
	{Tokens: []string{"better", "than", "expected"}, Weight: 2.0},
	{Tokens: []string{"buyback", "announced"}, Weight: 1.5},
	{Tokens: []string{"share", "buyback"}, Weight: 1.5},
	{Tokens: []string{"dividend", "increase"}, Weight: 1.5},
	{Tokens: []string{"dividend", "hike"}, Weight: 1.5},
	{Tokens: []string{"strong", "guidance"}, Weight: 2.0},
}

// bearishPhrases — checked before unigrams; longest-first ordering matters.
var bearishPhrases = []Phrase{
	{Tokens: []string{"filed", "for", "bankruptcy"}, Weight: -3.0},
	{Tokens: []string{"chapter", "11"}, Weight: -2.5},
	{Tokens: []string{"all", "time", "low"}, Weight: -2.5},
	{Tokens: []string{"all-time", "low"}, Weight: -2.5},
	{Tokens: []string{"record", "low"}, Weight: -2.0},
	{Tokens: []string{"cut", "guidance"}, Weight: -2.5},
	{Tokens: []string{"cuts", "guidance"}, Weight: -2.5},
	{Tokens: []string{"lowered", "guidance"}, Weight: -2.5},
	{Tokens: []string{"lowers", "guidance"}, Weight: -2.5},
	{Tokens: []string{"slashed", "guidance"}, Weight: -2.5},
	{Tokens: []string{"weak", "guidance"}, Weight: -2.0},
	{Tokens: []string{"missed", "estimates"}, Weight: -2.5},
	{Tokens: []string{"miss", "estimates"}, Weight: -2.0},
	{Tokens: []string{"misses", "estimates"}, Weight: -2.5},
	{Tokens: []string{"missed", "expectations"}, Weight: -2.0},
	{Tokens: []string{"below", "estimates"}, Weight: -2.0},
	{Tokens: []string{"earnings", "miss"}, Weight: -2.5},
	{Tokens: []string{"profit", "warning"}, Weight: -2.5},
	{Tokens: []string{"sec", "investigation"}, Weight: -2.0},
	{Tokens: []string{"class", "action"}, Weight: -1.5},
	{Tokens: []string{"product", "recall"}, Weight: -2.0},
	{Tokens: []string{"layoffs", "announced"}, Weight: -1.5},
	{Tokens: []string{"job", "cuts"}, Weight: -1.5},
	{Tokens: []string{"worse", "than", "expected"}, Weight: -2.0},
	{Tokens: []string{"weaker", "than", "expected"}, Weight: -2.0},
}

// bullishUnigrams — single tokens. Weights map → strong (2.0) or weak (1.0).
var bullishUnigrams = map[string]float64{
	// strong
	"surge": strongSignal, "surges": strongSignal, "surged": strongSignal, "surging": strongSignal,
	"soar": strongSignal, "soars": strongSignal, "soared": strongSignal, "soaring": strongSignal,
	"rally": strongSignal, "rallies": strongSignal, "rallied": strongSignal, "rallying": strongSignal,
	"jump": strongSignal, "jumps": strongSignal, "jumped": strongSignal,
	"beat": strongSignal, "beats": strongSignal,
	"outperform": strongSignal, "outperforms": strongSignal, "outperformed": strongSignal,
	"upgrade": strongSignal, "upgrades": strongSignal, "upgraded": strongSignal,
	"blockbuster": strongSignal, "blowout": strongSignal, "breakthrough": strongSignal,
	"exceeded": strongSignal, "exceeds": strongSignal,
	// moderate
	"gain": weakSignal, "gains": weakSignal, "gained": weakSignal, "gaining": weakSignal,
	"growth": weakSignal, "strong": weakSignal, "stronger": weakSignal,
	"positive": weakSignal, "up": weakSignal, "climb": weakSignal, "climbs": weakSignal, "climbed": weakSignal,
	"rise": weakSignal, "rises": weakSignal, "rose": weakSignal, "rising": weakSignal,
	"expansion": weakSignal, "profit": weakSignal, "profits": weakSignal, "profitable": weakSignal,
	"dividend": weakSignal, "buyback": weakSignal, "momentum": weakSignal,
	"bullish": weakSignal, "optimistic": weakSignal, "robust": weakSignal, "healthy": weakSignal,
	"boom": weakSignal, "booms": weakSignal,
}

// bearishUnigrams — single tokens. Stored as positive magnitudes; the scorer applies the sign.
var bearishUnigrams = map[string]float64{
	// strong
	"plunge": strongSignal, "plunges": strongSignal, "plunged": strongSignal, "plunging": strongSignal,
	"tumble": strongSignal, "tumbles": strongSignal, "tumbled": strongSignal, "tumbling": strongSignal,
	"crash": strongSignal, "crashes": strongSignal, "crashed": strongSignal, "crashing": strongSignal,
	"slump": strongSignal, "slumps": strongSignal, "slumped": strongSignal,
	"miss": strongSignal, "missed": strongSignal, "misses": strongSignal,
	"downgrade": strongSignal, "downgrades": strongSignal, "downgraded": strongSignal,
	"slash": strongSignal, "slashes": strongSignal, "slashed": strongSignal,
	"fraud": strongSignal, "bankruptcy": strongSignal, "lawsuit": strongSignal,
	"recall": strongSignal, "scandal": strongSignal, "probe": strongSignal,
	// moderate
	"drop": weakSignal, "drops": weakSignal, "dropped": weakSignal, "dropping": weakSignal,
	"fall": weakSignal, "falls": weakSignal, "fell": weakSignal, "falling": weakSignal,
	"loss": weakSignal, "losses": weakSignal, "decline": weakSignal, "declines": weakSignal, "declined": weakSignal,
	"weak": weakSignal, "weaker": weakSignal, "weakest": weakSignal,
	"negative": weakSignal, "down": weakSignal,
	"sell": weakSignal, "selloff": weakSignal, "sell-off": weakSignal,
	"underperform": weakSignal, "underperforms": weakSignal, "underperformed": weakSignal,
	"bearish": weakSignal, "concerns": weakSignal, "concern": weakSignal,
	"headwinds": weakSignal, "headwind": weakSignal,
	"sluggish": weakSignal, "soften": weakSignal, "softening": weakSignal,
	"struggling": weakSignal, "struggle": weakSignal, "slowdown": weakSignal, "slowing": weakSignal,
	"downturn": weakSignal,
}

// negators flip the sign of a polarity term that appears within negationWindow tokens.
var negators = map[string]struct{}{
	"not": {}, "no": {}, "without": {}, "never": {}, "neither": {}, "nor": {}, "none": {},
	"didn't": {}, "doesn't": {}, "don't": {}, "isn't": {}, "wasn't": {}, "weren't": {},
	"won't": {}, "wouldn't": {}, "can't": {}, "cannot": {}, "couldn't": {},
}

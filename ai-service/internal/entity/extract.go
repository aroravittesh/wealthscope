// Package entity implements stock-symbol and company-name extraction with
// position-aware ranking. Sources: $TICKER regex, plain ALL-CAPS regex, and a
// curated alias dictionary (see aliases.go). Primary entity is the earliest
// resolved symbol in the user's sentence; tier is a tiebreaker only.
package entity

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// EntityResult is the structured output of hybrid entity extraction.
// Shape is intentionally stable for downstream callers (intent, RAG, compare).
type EntityResult struct {
	PrimaryTicker    string
	SecondaryTickers []string
	CompanyMatches   []string
	Confidence       float64
}

// Tiers (higher = stronger signal). Used only to break ties between two
// candidates that share a starting position.
const (
	tierDict   = 1
	tierPlain  = 2
	tierDollar = 3
)

// candidate is one extracted entity occurrence.
type candidate struct {
	symbol string
	tier   int
	pos    int    // byte offset of the match in the original message
	label  string // canonical company label, "" if from ticker scan
}

// ignorePlainTicker tokens are never treated as tickers when matched as plain
// uppercase words. Keep this conservative — real tickers like SPY, ETF holdings
// names, etc. should pass through.
var ignorePlainTicker = map[string]bool{
	// English connectives / pronouns
	"I": true, "A": true, "THE": true, "AND": true, "OR": true,
	"FOR": true, "IN": true, "IS": true, "IT": true, "ON": true,
	"AT": true, "TO": true, "DO": true, "GO": true, "BE": true,
	"BY": true, "AN": true, "UP": true, "IF": true, "NO": true,
	"SO": true, "MY": true, "US": true, "WE": true, "ME": true,
	"OF": true, "AS": true, "AM": true, "OK": true, "OH": true,
	// Question / prose words often shouted in caps
	"WHAT": true, "HOW": true, "WHY": true, "WHO": true, "WHEN": true,
	"WHERE": true, "WHICH": true, "DOING": true, "TODAY": true, "ABOUT": true,
	// Finance jargon that looks like tickers but isn't tradable as such
	"AI": true, "ML": true, "IPO": true, "CEO": true, "CFO": true,
	"COO": true, "CTO": true, "EPS": true, "ETF": true, "ESG": true,
	"ROI": true, "ROE": true, "USD": true, "EUR": true, "GBP": true,
	"JPY": true, "CNY": true, "GDP": true, "USA": true, "UK": true,
	"EU": true, "FED": true, "NEWS": true, "STOCK": true, "SHARES": true,
	"NYSE": true, "SEC": true, "GMT": true, "EST": true, "PST": true,
	"COVID": true, "FAANG": true, "MAANG": true, "EBITDA": true,
	"MARKET": true, "STOCKS": true, "PRICE": true, "PRICES": true,
	"BUY": true, "SELL": true, "HOLD": true, "RISK": true, "BULL": true,
	"BEAR": true, "ALL": true, "TIME": true, "HIGH": true, "LOW": true,
}

var (
	reDollarTicker = regexp.MustCompile(`\$([A-Z][A-Z.]{0,5})\b`)
	rePlainTicker  = regexp.MustCompile(`\b([A-Z]{2,5})\b`)
)

// Extract runs hybrid extraction: $ tickers, plain tickers, then alias matches.
// Candidates are ordered by char position (sentence order); ties are broken by
// tier so a stronger signal at the same offset wins.
func Extract(message string) EntityResult {
	if strings.TrimSpace(message) == "" {
		return EntityResult{}
	}

	var cands []candidate
	cands = append(cands, scanDollarTickers(message)...)

	if !looksLikeAllCapsProse(message) {
		cands = append(cands, scanPlainTickers(message)...)
	}

	cands = append(cands, scanCompanyAliases(message)...)

	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].pos != cands[j].pos {
			return cands[i].pos < cands[j].pos
		}
		return cands[i].tier > cands[j].tier
	})

	uniqueSyms, labels := dedupeCandidates(cands)
	if len(uniqueSyms) == 0 {
		return EntityResult{}
	}

	primary := uniqueSyms[0]
	var secondary []string
	if len(uniqueSyms) > 1 {
		secondary = append([]string{}, uniqueSyms[1:]...)
	}

	// Confidence is driven by the strongest source confirming the primary
	// (so "Apple ($AAPL)" benefits from the dollar-ticker tier even though
	// "Apple" appears first in the sentence).
	conf := baseConfidence(strongestTierFor(cands, primary))
	if hasMultiSourceAgreement(cands, primary) {
		conf = clamp01(conf + 0.05)
	}

	return EntityResult{
		PrimaryTicker:    primary,
		SecondaryTickers: secondary,
		CompanyMatches:   labels,
		Confidence:       conf,
	}
}

func scanDollarTickers(message string) []candidate {
	out := make([]candidate, 0)
	for _, m := range reDollarTicker.FindAllStringSubmatchIndex(message, -1) {
		if len(m) < 4 {
			continue
		}
		sym := message[m[2]:m[3]]
		out = append(out, candidate{symbol: sym, tier: tierDollar, pos: m[0]})
	}
	return out
}

func scanPlainTickers(message string) []candidate {
	out := make([]candidate, 0)
	for _, m := range rePlainTicker.FindAllStringSubmatchIndex(message, -1) {
		if len(m) < 4 {
			continue
		}
		sym := message[m[2]:m[3]]
		if ignorePlainTicker[sym] {
			continue
		}
		out = append(out, candidate{symbol: sym, tier: tierPlain, pos: m[0]})
	}
	return out
}

// scanCompanyAliases matches alias phrases against a normalized message
// (lowercase, hyphens → spaces). Positions are byte offsets in the normalized
// form; since both transforms are length-preserving, they map 1:1 to the
// original string for ranking purposes.
func scanCompanyAliases(message string) []candidate {
	norm := normalizeForAlias(message)
	consumed := make([]bool, len(norm))
	out := make([]candidate, 0)
	for _, m := range compiledAliases {
		for _, idx := range m.re.FindAllStringIndex(norm, -1) {
			start, end := idx[0], idx[1]
			if anyConsumedRange(consumed, start, end) {
				continue
			}
			markConsumedRange(consumed, start, end)
			out = append(out, candidate{
				symbol: m.alias.Ticker,
				tier:   tierDict,
				pos:    start,
				label:  m.alias.Label,
			})
		}
	}
	return out
}

func dedupeCandidates(cands []candidate) (symbols []string, labels []string) {
	if len(cands) == 0 {
		return nil, nil
	}
	seenSym := map[string]bool{}
	seenLbl := map[string]bool{}
	for _, c := range cands {
		if !seenSym[c.symbol] {
			seenSym[c.symbol] = true
			symbols = append(symbols, c.symbol)
		}
		if c.label != "" && !seenLbl[c.label] {
			seenLbl[c.label] = true
			labels = append(labels, c.label)
		}
	}
	return symbols, labels
}

// strongestTierFor returns the highest-tier candidate matching the given symbol.
// Used so confidence reflects the most reliable source confirming the primary.
func strongestTierFor(cands []candidate, symbol string) int {
	best := 0
	for _, c := range cands {
		if c.symbol == symbol && c.tier > best {
			best = c.tier
		}
	}
	return best
}

// hasMultiSourceAgreement reports whether the resolved primary ticker was
// produced by two or more distinct extraction sources (dollar / plain / dict).
// Used to award a small confidence bonus when signals corroborate each other.
func hasMultiSourceAgreement(cands []candidate, primary string) bool {
	tiers := map[int]bool{}
	for _, c := range cands {
		if c.symbol == primary {
			tiers[c.tier] = true
		}
	}
	return len(tiers) >= 2
}

func baseConfidence(primaryTier int) float64 {
	switch primaryTier {
	case tierDollar:
		return 0.95
	case tierPlain:
		return 0.85
	case tierDict:
		return 0.75
	default:
		return 0
	}
}

// looksLikeAllCapsProse returns true for messages dominated by uppercase
// letters (e.g., shouted prose). Such inputs are excluded from the plain
// uppercase ticker regex to prevent words like THE/MARKET being misread.
func looksLikeAllCapsProse(message string) bool {
	letters, upper := 0, 0
	for _, r := range message {
		if !unicode.IsLetter(r) {
			continue
		}
		letters++
		if unicode.IsUpper(r) {
			upper++
		}
	}
	if letters < 8 {
		return false
	}
	return float64(upper)/float64(letters) > 0.7
}

// normalizeForAlias lowercases and replaces hyphens with spaces. The output
// has the same byte length as the input so positions remain comparable.
func normalizeForAlias(s string) string {
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '-':
			b.WriteRune(' ')
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func anyConsumedRange(consumed []bool, start, end int) bool {
	if end > len(consumed) {
		end = len(consumed)
	}
	for i := start; i < end; i++ {
		if consumed[i] {
			return true
		}
	}
	return false
}

func markConsumedRange(consumed []bool, start, end int) {
	if end > len(consumed) {
		end = len(consumed)
	}
	for i := start; i < end; i++ {
		consumed[i] = true
	}
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

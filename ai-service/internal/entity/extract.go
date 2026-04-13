package entity

import (
	"regexp"
	"strings"
	"unicode"
)

// EntityResult is the structured output of hybrid entity extraction.
type EntityResult struct {
	PrimaryTicker    string
	SecondaryTickers []string
	CompanyMatches   []string
	Confidence       float64
}

// companyToTicker maps normalized company name tokens to tickers.
var companyToTicker = map[string]string{
	"apple":     "AAPL",
	"microsoft": "MSFT",
	"tesla":     "TSLA",
	"nvidia":    "NVDA",
	"google":    "GOOGL",
	"alphabet":  "GOOGL",
	"amazon":    "AMZN",
	"meta":      "META",
	"facebook":  "META",
	"netflix":   "NFLX",
}

// canonicalCompanyLabel is a short display name for CompanyMatches.
var canonicalCompanyLabel = map[string]string{
	"apple":     "Apple",
	"microsoft": "Microsoft",
	"tesla":     "Tesla",
	"nvidia":    "NVIDIA",
	"google":    "Google",
	"alphabet":  "Alphabet",
	"amazon":    "Amazon",
	"meta":      "Meta",
	"facebook":  "Facebook",
	"netflix":   "Netflix",
}

// ignorePlainTicker tokens are not treated as tickers when matched as plain uppercase words.
var ignorePlainTicker = map[string]bool{
	"I": true, "A": true, "THE": true, "AND": true, "OR": true,
	"FOR": true, "IN": true, "IS": true, "IT": true, "ON": true,
	"AT": true, "TO": true, "DO": true, "GO": true, "BE": true,
	"BY": true, "AN": true, "UP": true, "IF": true, "NO": true,
	"SO": true, "MY": true, "US": true, "WE": true,
	// prose / finance questions often shouted in all caps
	"MARKET": true, "WHAT": true, "HOW": true, "WHY": true,
	"DOING": true, "TODAY": true, "ABOUT": true,
}

var (
	reDollarTicker = regexp.MustCompile(`\$([A-Z]{1,5})\b`)
	rePlainTicker  = regexp.MustCompile(`\b([A-Z]{1,5})\b`)
)

// Extract runs hybrid extraction: $ tickers, plain tickers, then company names.
func Extract(message string) EntityResult {
	var dollarSyms []string
	for _, m := range reDollarTicker.FindAllStringSubmatch(message, -1) {
		if len(m) > 1 {
			dollarSyms = append(dollarSyms, m[1])
		}
	}

	var plainSyms []string
	for _, m := range rePlainTicker.FindAllStringSubmatch(message, -1) {
		if len(m) < 2 {
			continue
		}
		s := m[1]
		if ignorePlainTicker[s] {
			continue
		}
		plainSyms = append(plainSyms, s)
	}

	var dictSyms []string
	var companyNames []string
	for _, tok := range wordTokens(message) {
		if sym, ok := companyToTicker[tok]; ok {
			dictSyms = append(dictSyms, sym)
			if label, ok := canonicalCompanyLabel[tok]; ok {
				companyNames = append(companyNames, label)
			}
		}
	}

	ordered := mergeUnique(dollarSyms, plainSyms, dictSyms)
	companyNames = dedupeStrings(companyNames)

	conf := confidence(dollarSyms, plainSyms, len(ordered))

	if len(ordered) == 0 {
		return EntityResult{Confidence: conf}
	}

	second := []string(nil)
	if len(ordered) > 1 {
		second = append(second, ordered[1:]...)
	}
	return EntityResult{
		PrimaryTicker:    ordered[0],
		SecondaryTickers: second,
		CompanyMatches:   companyNames,
		Confidence:       conf,
	}
}

func mergeUnique(dollar, plain, dict []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, tier := range [][]string{dollar, plain, dict} {
		for _, s := range tier {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func wordTokens(s string) []string {
	s = strings.ToLower(s)
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		tokens = append(tokens, cur.String())
		cur.Reset()
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func confidence(dollar, plain []string, n int) float64 {
	if n == 0 {
		return 0
	}
	switch {
	case len(dollar) > 0:
		return 0.95
	case len(plain) > 0:
		return 0.85
	default:
		return 0.75
	}
}

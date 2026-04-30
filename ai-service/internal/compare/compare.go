package compare

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"wealthscope-ai/internal/market"
)

var (
	ErrInvalidSymbolCount = errors.New("symbols must include between 2 and 4 non-empty unique tickers")
	ErrEmptySymbol        = errors.New("symbols cannot be empty or whitespace")
)

// Request is the JSON body for POST /compare.
type Request struct {
	Symbols []string `json:"symbols"`
}

// StockComparison is one row in the compare response.
type StockComparison struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Beta      string `json:"beta"`
	Sector    string `json:"sector"`
	Industry  string `json:"industry"`
	MarketCap string `json:"market_cap"`
	PERatio   string `json:"pe_ratio"`
}

// Response is the JSON output for POST /compare.
type Response struct {
	Comparisons []StockComparison `json:"comparisons"`
	Summary     string            `json:"summary"`
}

// Fetcher loads quote and overview per symbol (live implementation uses market package).
type Fetcher interface {
	Quote(symbol string) (*market.GlobalQuote, error)
	Overview(symbol string) (*market.CompanyOverview, error)
}

// LiveFetcher uses production market API helpers.
type LiveFetcher struct{}

func (LiveFetcher) Quote(symbol string) (*market.GlobalQuote, error) {
	return market.GetStockQuote(symbol)
}

func (LiveFetcher) Overview(symbol string) (*market.CompanyOverview, error) {
	return market.GetCompanyOverview(symbol)
}

// NormalizeAndValidate returns uppercase unique symbols; count must be 2–4.
func NormalizeAndValidate(raw []string) ([]string, error) {
	if len(raw) < 2 || len(raw) > 4 {
		return nil, ErrInvalidSymbolCount
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(strings.ToUpper(s))
		if s == "" {
			return nil, ErrEmptySymbol
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) < 2 || len(out) > 4 {
		return nil, ErrInvalidSymbolCount
	}
	return out, nil
}

// Compare loads data for each symbol and builds a neutral summary (no recommendations).
func Compare(f Fetcher, symbols []string) (*Response, error) {
	norm, err := NormalizeAndValidate(symbols)
	if err != nil {
		return nil, err
	}

	rows := make([]StockComparison, 0, len(norm))
	for _, sym := range norm {
		q, err := f.Quote(sym)
		if err != nil {
			return nil, fmt.Errorf("%s: quote: %w", sym, err)
		}
		o, err := f.Overview(sym)
		if err != nil {
			return nil, fmt.Errorf("%s: overview: %w", sym, err)
		}
		rows = append(rows, StockComparison{
			Symbol:    sym,
			Price:     q.Price,
			Beta:      o.Beta,
			Sector:    o.Sector,
			Industry:  o.Industry,
			MarketCap: o.MarketCap,
			PERatio:   o.PERatio,
		})
	}

	return &Response{
		Comparisons: rows,
		Summary:     buildSummary(rows),
	}, nil
}

func buildSummary(rows []StockComparison) string {
	if len(rows) == 0 {
		return ""
	}

	names := make([]string, len(rows))
	sectors := make(map[string]struct{})
	for i, r := range rows {
		names[i] = r.Symbol
		if r.Sector != "" {
			sectors[r.Sector] = struct{}{}
		}
	}

	var b strings.Builder
	b.WriteString("Side-by-side snapshot for ")
	b.WriteString(strings.Join(names, ", "))
	b.WriteString(". ")

	if len(sectors) == 1 {
		for s := range sectors {
			b.WriteString("All names are in the same sector (" + s + "). ")
			break
		}
	} else if len(sectors) > 1 {
		secList := make([]string, 0, len(sectors))
		for s := range sectors {
			secList = append(secList, s)
		}
		b.WriteString("Sectors represented: " + strings.Join(secList, ", ") + ". ")
	}

	lowB, highB, lowSym, highSym := betaRange(rows)
	if lowB != highB && lowSym != "" && highSym != "" {
		b.WriteString(fmt.Sprintf(
			"Reported beta is lowest for %s (%.2f) and highest for %s (%.2f), so market sensitivity differs across these names. ",
			lowSym, lowB, highSym, highB,
		))
	} else if lowSym != "" {
		b.WriteString(fmt.Sprintf("Reported beta is around %.2f across the set (market sensitivity looks similar in this snapshot). ", lowB))
	}

	lowPE, highPE, lowPESym, highPESym := peRange(rows)
	if lowPESym != "" && highPESym != "" && lowPE != highPE {
		b.WriteString(fmt.Sprintf(
			"P/E ratios range from about %.2f (%s) to %.2f (%s), so valuation multiples are not uniform. ",
			lowPE, lowPESym, highPE, highPESym,
		))
	}

	lowCap, highCap, lowCapSym, highCapSym := capRange(rows)
	if lowCapSym != "" && highCapSym != "" && lowCap != highCap {
		b.WriteString(fmt.Sprintf(
			"Market capitalization (as reported) is larger for %s than for %s in this comparison. ",
			highCapSym, lowCapSym,
		))
	}

	b.WriteString("Figures are data snapshots only; they are not buy or sell signals.")

	return strings.TrimSpace(b.String())
}

func betaRange(rows []StockComparison) (low, high float64, lowSym, highSym string) {
	low = 1e9
	high = -1e9
	for _, r := range rows {
		v, ok := parseFloat(r.Beta)
		if !ok {
			continue
		}
		if v < low {
			low, lowSym = v, r.Symbol
		}
		if v > high {
			high, highSym = v, r.Symbol
		}
	}
	if high < -1e8 {
		return 0, 0, "", ""
	}
	return low, high, lowSym, highSym
}

func peRange(rows []StockComparison) (low, high float64, lowSym, highSym string) {
	low = 1e9
	high = -1e9
	for _, r := range rows {
		v, ok := parseFloat(r.PERatio)
		if !ok {
			continue
		}
		if v < low {
			low, lowSym = v, r.Symbol
		}
		if v > high {
			high, highSym = v, r.Symbol
		}
	}
	if high < -1e8 {
		return 0, 0, "", ""
	}
	return low, high, lowSym, highSym
}

func capRange(rows []StockComparison) (low, high float64, lowSym, highSym string) {
	low = 1e30
	high = -1.0
	for _, r := range rows {
		v, ok := parseFloat(r.MarketCap)
		if !ok {
			continue
		}
		if v < low {
			low, lowSym = v, r.Symbol
		}
		if v > high {
			high, highSym = v, r.Symbol
		}
	}
	if high < 0 {
		return 0, 0, "", ""
	}
	return low, high, lowSym, highSym
}

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "none") || s == "-" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

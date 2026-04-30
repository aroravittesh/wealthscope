// Package websearch is the backend-controlled web/news search layer used by
// the chat pipeline to ground answers about time-sensitive market topics.
//
// The package is intentionally provider-agnostic: a Provider interface lets the
// real Tavily client be swapped for a stub or mock during tests or for any
// future provider (SerpAPI, Bing News, Brave Search, etc.).
package websearch

import (
	"context"
	"strings"
)

// Provider is the abstraction over a structured web/news search API. Returns
// at most maxResults raw results; callers run CleanAndRank afterwards.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string, maxResults int) ([]Result, error)
}

// Result is one search hit. Snippet is a short text body suitable for prompt
// grounding; URL is preserved internally for diagnostics/audit but is not sent
// to the LLM verbatim.
type Result struct {
	Title       string  `json:"title"`
	Snippet     string  `json:"snippet"`
	URL         string  `json:"url,omitempty"`
	Source      string  `json:"source"`
	PublishedAt string  `json:"published_at,omitempty"`
	Score       float64 `json:"score,omitempty"`
}

// Decision is the output of Decide. Use=true means the chat pipeline should
// call the configured provider with the supplied Query; Reason is a short
// explanation suitable for logging.
type Decision struct {
	Use    bool
	Reason string
	Query  string
}

// trustedDomains is the curated list of finance/news outlets we boost during
// ranking. Domains are matched case-insensitively against result.Source or the
// hostname of result.URL.
var trustedDomains = map[string]struct{}{
	"reuters.com":         {},
	"bloomberg.com":       {},
	"wsj.com":             {},
	"ft.com":              {},
	"cnbc.com":            {},
	"marketwatch.com":     {},
	"finance.yahoo.com":   {},
	"yahoo.com":           {},
	"sec.gov":             {},
	"federalreserve.gov":  {},
	"investopedia.com":    {},
	"businessinsider.com": {},
	"barrons.com":         {},
	"seekingalpha.com":    {},
	"fortune.com":         {},
	"axios.com":           {},
	"forbes.com":          {},
	"morningstar.com":     {},
	"nasdaq.com":          {},
}

// IsTrustedSource reports whether s matches a curated finance/news domain.
func IsTrustedSource(s string) bool {
	host := strings.ToLower(strings.TrimSpace(s))
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return false
	}
	if _, ok := trustedDomains[host]; ok {
		return true
	}
	// Allow subdomains like "feeds.reuters.com" → "reuters.com".
	for d := range trustedDomains {
		if strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

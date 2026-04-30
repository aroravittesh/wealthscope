package websearch

import (
	"net/url"
	"sort"
	"strings"
)

// Cap on individual snippet length sent into the prompt. Longer snippets are
// truncated; we never reject for length alone.
const maxSnippetChars = 280

// Minimum useful snippet length. Anything shorter is dropped because the LLM
// gains very little from a 1-line teaser.
const minSnippetChars = 30

// Two-tier ranking: any result from a trusted finance/news domain is ordered
// before any non-trusted result. Within each tier, provider score (descending)
// breaks ties. This is a deliberately decisive trust policy — when WealthScope
// is grounding "latest news" claims for users, we prefer a marginal Reuters
// hit over a high-scoring random domain.

// CleanAndRank applies safety / quality controls and returns at most topK
// results sorted by (boost + score) descending.
//
// Filtering rules:
//   - drop empty title/snippet
//   - drop snippets shorter than minSnippetChars
//   - dedupe by URL (case-insensitive)
//   - dedupe by (host, lower-cased title prefix)
//   - prefer trusted domains via score boost
//   - truncate snippet to maxSnippetChars
func CleanAndRank(in []Result, topK int) []Result {
	if topK <= 0 {
		topK = 3
	}
	if len(in) == 0 {
		return nil
	}

	seenURL := make(map[string]struct{}, len(in))
	seenHostTitle := make(map[string]struct{}, len(in))
	out := make([]Result, 0, len(in))

	for _, r := range in {
		title := strings.TrimSpace(r.Title)
		snippet := strings.TrimSpace(r.Snippet)
		if title == "" || snippet == "" {
			continue
		}
		if len([]rune(snippet)) < minSnippetChars {
			continue
		}

		host := hostnameOf(r.URL, r.Source)
		canonURL := strings.ToLower(strings.TrimSpace(r.URL))
		if canonURL != "" {
			if _, dup := seenURL[canonURL]; dup {
				continue
			}
			seenURL[canonURL] = struct{}{}
		}
		key := host + "|" + lowerPrefix(title, 60)
		if _, dup := seenHostTitle[key]; dup {
			continue
		}
		seenHostTitle[key] = struct{}{}

		cleaned := r
		cleaned.Title = title
		cleaned.Snippet = truncateRunes(snippet, maxSnippetChars)
		if cleaned.Source == "" {
			cleaned.Source = host
		}
		out = append(out, cleaned)
	}

	sort.SliceStable(out, func(i, j int) bool {
		ti := isTrustedResult(out[i])
		tj := isTrustedResult(out[j])
		if ti != tj {
			return ti // trusted bucket wins
		}
		return out[i].Score > out[j].Score
	})

	if len(out) > topK {
		out = out[:topK]
	}
	return out
}

func isTrustedResult(r Result) bool {
	if IsTrustedSource(r.Source) {
		return true
	}
	return IsTrustedSource(hostnameOf(r.URL, r.Source))
}

func hostnameOf(rawURL, fallback string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL != "" {
		if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
			return strings.ToLower(strings.TrimPrefix(u.Host, "www."))
		}
	}
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(fallback), "www."))
}

func lowerPrefix(s string, n int) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

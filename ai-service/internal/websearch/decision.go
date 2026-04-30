package websearch

import (
	"strings"

	"wealthscope-ai/internal/entity"
)

// Time-sensitive trigger tokens. A query needs at least one to qualify for a
// live web call. Tokens are lower-case substrings; matching is case-insensitive.
var timeSensitiveTokens = []string{
	"latest", "today", "recent", "recently", "news", "update", "updates",
	"happening", "happened", "this week", "this month", "current",
	"just", "breaking", "now", "lately", "yesterday", "this morning",
}

// Phrases that should always force a web search for time-sensitive grounding,
// regardless of the classified intent. They strongly indicate the user wants
// a current event answer.
var explicitWebPhrases = []string{
	"news on", "news about", "what happened to", "what happened with",
	"what's happening", "whats happening", "what is happening",
	"any news", "what's new", "whats new",
}

// Phrases that look like evergreen finance questions. If a query matches one
// of these AND has no time-sensitive token AND no entity, we skip web search
// because internal RAG/KB is the right surface.
var evergreenPrefixes = []string{
	"what is ", "what are ", "define ", "explain ", "how does ", "how do ",
	"meaning of ", "definition of ",
}

// Intents for which web search may be helpful. Intents like PORTFOLIO_TIP are
// excluded — those are handled by dedicated explain endpoints.
var webEligibleIntents = map[string]struct{}{
	"STOCK_PRICE":    {},
	"MARKET_NEWS":    {},
	"GENERAL_MARKET": {},
	"RISK_ANALYSIS":  {},
	// UNKNOWN is allowed so that a clearly time-sensitive freeform question
	// ("any news on apple today?") still triggers when the classifier is
	// uncertain. Decide() still requires a time-sensitive token in that case.
	"UNKNOWN": {},
}

// Decide evaluates whether the chat pipeline should perform a live web search
// for this turn. It is deterministic, depends only on the inputs, and returns
// a short Reason string for logging/explainability.
//
// Rules:
//   - Intent must be web-eligible.
//   - Query must contain at least one time-sensitive token, OR an explicit
//     news phrase.
//   - Evergreen definitional questions with no time token and no entity are
//     skipped — internal KB is preferred.
//   - When triggered, Query is built as "<TICKER> stock latest news" if a
//     primary ticker is known, otherwise the trimmed user message.
func Decide(message string, intent string, ent entity.EntityResult) Decision {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return Decision{Use: false, Reason: "empty message"}
	}
	intent = strings.ToUpper(strings.TrimSpace(intent))

	hasTimeToken := containsAny(lower, timeSensitiveTokens)
	hasExplicit := containsAny(lower, explicitWebPhrases)

	// Evergreen guard: only when there is no time signal and no entity.
	if !hasTimeToken && !hasExplicit && ent.PrimaryTicker == "" && hasEvergreenPrefix(lower) {
		return Decision{Use: false, Reason: "evergreen definition; internal KB preferred"}
	}

	if !hasTimeToken && !hasExplicit {
		return Decision{Use: false, Reason: "no time-sensitive cue"}
	}

	if _, ok := webEligibleIntents[intent]; !ok {
		// PORTFOLIO_TIP and similar intents stay internal-only.
		return Decision{Use: false, Reason: "intent " + intent + " is not web-eligible"}
	}

	q := buildQuery(message, ent)
	reason := "time-sensitive query"
	if hasExplicit {
		reason = "explicit news phrase"
	}
	if ent.PrimaryTicker != "" {
		reason = reason + " for " + ent.PrimaryTicker
	}
	return Decision{Use: true, Reason: reason, Query: q}
}

func buildQuery(message string, ent entity.EntityResult) string {
	if t := strings.TrimSpace(ent.PrimaryTicker); t != "" {
		// Tickers alone make for ambiguous web queries (e.g. "T" → AT&T plus
		// a thousand T-prefixed false positives). Append "stock latest news"
		// so the search is biased toward the right surface.
		return t + " stock latest news"
	}
	q := strings.TrimSpace(message)
	const cap = 200
	if len(q) > cap {
		q = q[:cap]
	}
	return q
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

func hasEvergreenPrefix(lower string) bool {
	for _, p := range evergreenPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

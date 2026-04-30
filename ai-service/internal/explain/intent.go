package explain

import (
	"fmt"
	"sort"
	"strings"
)

// Intent source labels are duplicated here as constants so callers (e.g. the
// frontend) get a canonical vocabulary without importing internal/ml.
const (
	IntentSourceRemote          = "remote_classifier"
	IntentSourceRemoteFallback  = "remote_fallback_keyword"
	IntentSourceLowConfFallback = "low_confidence_fallback"
	IntentSourceKeyword         = "keyword_only"
)

// BuildIntentExplanation constructs a structured explanation for an intent
// classification call.
//
//   - intent / ticker / source / confidence: the finalised classification.
//   - matchedKeywords: keyword-scorer hits when applicable; nil for remote.
//
// The returned Explanation is safe to JSON-marshal directly onto the API
// response struct.
func BuildIntentExplanation(intent, ticker, source string, confidence float64, matchedKeywords []string) Explanation {
	reasons := make([]string, 0, 4)
	signals := make([]Signal, 0, len(matchedKeywords)+1)

	switch source {
	case IntentSourceRemote:
		reasons = append(reasons,
			fmt.Sprintf("Trained intent classifier returned %s with confidence %.2f.", intent, confidence))
		signals = append(signals, Signal{
			Code:   "INTENT_REMOTE",
			Label:  "trained classifier",
			Score:  confidence,
			Detail: "TF-IDF + Logistic Regression remote inference.",
		})
	case IntentSourceLowConfFallback:
		reasons = append(reasons,
			fmt.Sprintf("Trained classifier confidence (%.2f) was below the configured threshold; using keyword scorer instead.", confidence))
	case IntentSourceRemoteFallback:
		reasons = append(reasons,
			"Trained classifier was unreachable or returned an invalid response; using keyword scorer instead.")
	case IntentSourceKeyword:
		reasons = append(reasons, "Trained classifier is not configured; using keyword scorer.")
	}

	if ticker != "" {
		reasons = append(reasons, fmt.Sprintf("Entity extractor identified primary ticker %s.", ticker))
		signals = append(signals, Signal{
			Code:   "ENTITY_TICKER",
			Label:  ticker,
			Score:  1.0,
			Detail: "Primary ticker identified by the entity resolver.",
		})
	}

	if len(matchedKeywords) > 0 {
		kws := append([]string(nil), matchedKeywords...)
		sort.Strings(kws)
		for _, kw := range kws {
			signals = append(signals, Signal{
				Code:   "INTENT_KEYWORD",
				Label:  kw,
				Score:  1.0,
				Detail: fmt.Sprintf("Phrase %q matched the %s lexicon.", kw, intent),
			})
		}
		reasons = append(reasons, fmt.Sprintf("Keyword scorer matched: %s.", strings.Join(kws, ", ")))
	}

	return Explanation{
		Code:       intentReasonCode(source, intent),
		Summary:    buildIntentSummary(intent, ticker, source, len(matchedKeywords)),
		Confidence: confidence,
		Source:     source,
		Reasons:    reasons,
		TopSignals: signals,
	}
}

func intentReasonCode(source, intent string) string {
	switch source {
	case IntentSourceRemote:
		return "INTENT_REMOTE_HIGH_CONFIDENCE"
	case IntentSourceLowConfFallback:
		return "INTENT_REMOTE_LOW_CONFIDENCE_FALLBACK"
	case IntentSourceRemoteFallback:
		return "INTENT_REMOTE_ERROR_FALLBACK"
	case IntentSourceKeyword:
		if intent == "UNKNOWN" {
			return "INTENT_KEYWORD_NO_MATCH"
		}
		return "INTENT_KEYWORD_MATCH"
	default:
		return "INTENT_UNKNOWN_SOURCE"
	}
}

func buildIntentSummary(intent, ticker, source string, kwHits int) string {
	parts := []string{fmt.Sprintf("Classified as %s", intent)}
	switch source {
	case IntentSourceRemote:
		parts = append(parts, "via the trained TF-IDF classifier")
	case IntentSourceLowConfFallback:
		parts = append(parts, "via the keyword scorer (trained classifier confidence too low)")
	case IntentSourceRemoteFallback:
		parts = append(parts, "via the keyword fallback (trained classifier unavailable)")
	case IntentSourceKeyword:
		parts = append(parts, "via the keyword scorer")
	}
	if kwHits > 0 {
		parts = append(parts, fmt.Sprintf("with %d matched keyword phrase(s)", kwHits))
	}
	if ticker != "" {
		parts = append(parts, fmt.Sprintf("and ticker %s detected", ticker))
	}
	return strings.Join(parts, " ") + "."
}

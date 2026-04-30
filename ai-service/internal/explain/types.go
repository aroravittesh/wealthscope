// Package explain hosts the shared explainability schema used across the AI
// service (intent classification, sentiment, portfolio risk / drift, RAG
// retrieval). The goal is a single, predictable structure every output can
// attach so the chatbot, the dashboard, and operators all see the *why*
// behind a decision in the same shape.
//
// Design notes:
//   - Stays a pure leaf package: no upstream imports back into ml/prediction,
//     so it can never create import cycles.
//   - Only depends on small leaf packages (portfoliorisk, finsentiment) for
//     convenience converters — those packages do not import explain.
//   - Every field is omitempty in the consuming response structs so older
//     clients keep working without changes.
package explain

// Signal is one machine-readable piece of evidence behind a decision.
//
// Examples by Code:
//   - INTENT_KEYWORD              keyword scorer match for intent routing
//   - INTENT_REMOTE               trained classifier prediction signal
//   - ENTITY_TICKER               primary ticker resolved by entity layer
//   - FIN_TERM_BULLISH/BEARISH    finance lexicon unigram hit
//   - FIN_PHRASE_BULLISH/BEARISH  finance lexicon multi-word phrase hit
//   - RISK_DRIVER_<CODE>          portfolio risk driver (BETA, HHI, ...)
//   - DRIFT_DRIVER_<CODE>         portfolio drift driver
//   - RAG_SEMANTIC/LEXICAL/...    retrieval scoring components
type Signal struct {
	Code     string  `json:"code"`
	Label    string  `json:"label"`
	Score    float64 `json:"score"`
	Polarity float64 `json:"polarity,omitempty"`
	Detail   string  `json:"detail,omitempty"`
}

// Explanation is the standard explainability envelope every output produces.
//
//   - Code is a stable machine identifier suitable for analytics and routing.
//   - Summary is a 1-2 sentence user-facing reason; safe to render in the UI.
//   - Reasons are short bullets suitable for an expandable "why" panel.
//   - TopSignals is the structured evidence behind the decision.
//   - Disclaimer is included on outputs that touch financial decisions.
type Explanation struct {
	Code       string   `json:"code"`
	Summary    string   `json:"summary"`
	Confidence float64  `json:"confidence,omitempty"`
	Source     string   `json:"source,omitempty"`
	Reasons    []string `json:"reasons,omitempty"`
	TopSignals []Signal `json:"top_signals,omitempty"`
	Disclaimer string   `json:"disclaimer,omitempty"`
}

// EducationalDisclaimer is the standard footer for finance-facing outputs.
const EducationalDisclaimer = "Educational signal only. This is not financial advice."

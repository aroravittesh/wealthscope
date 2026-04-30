// Package feedback implements the v1 feedback collection pipeline.
//
// This is a learning-oriented logging system, not an online-learning loop.
// Records are validated, sanitised, and appended to a JSONL log so they can
// be loaded into a notebook (pandas read_json(..., lines=True)) or shipped
// to a warehouse later for retraining datasets.
package feedback

import (
	"errors"
	"strings"
	"time"
)

// Allowed feedback labels. Keep this list small and stable so downstream
// analytics and retraining jobs can rely on it.
const (
	LabelHelpful       = "helpful"
	LabelNotHelpful    = "not_helpful"
	LabelIncorrect     = "incorrect"
	LabelUnclear       = "unclear"
	LabelIrrelevant    = "irrelevant"
	LabelUnsafe        = "unsafe"
	LabelPoorRetrieval = "poor_retrieval"
)

// Allowed response_type values. Each maps to one or more product surfaces.
const (
	ResponseTypeChat       = "chat"
	ResponseTypeRisk       = "risk"
	ResponseTypeDrift      = "drift"
	ResponseTypeSentiment  = "sentiment"
	ResponseTypeRetrieval  = "retrieval"
	ResponseTypeSummarize  = "summarize"
	ResponseTypeExplain    = "explain"
	ResponseTypeCompare    = "compare"
	ResponseTypeIntent     = "intent"
)

// AllowedLabels is the canonical set of accepted feedback labels.
var AllowedLabels = []string{
	LabelHelpful,
	LabelNotHelpful,
	LabelIncorrect,
	LabelUnclear,
	LabelIrrelevant,
	LabelUnsafe,
	LabelPoorRetrieval,
}

// AllowedResponseTypes is the canonical set of accepted response_type values.
var AllowedResponseTypes = []string{
	ResponseTypeChat,
	ResponseTypeRisk,
	ResponseTypeDrift,
	ResponseTypeSentiment,
	ResponseTypeRetrieval,
	ResponseTypeSummarize,
	ResponseTypeExplain,
	ResponseTypeCompare,
	ResponseTypeIntent,
}

// Field-size caps. Truncation (not rejection) keeps long pastes from blowing
// up the log while still preserving the head of the message for analysis.
const (
	maxSessionIDLen    = 128
	maxMessageIDLen    = 128
	maxQueryLen        = 1000
	maxReasonLen       = 1000
	maxResponseExcerpt = 500
	maxEndpointLen     = 128
	maxDocIDLen        = 64
	maxDocIDs          = 16
)

// Feedback is one user-feedback record. Server fills ID and Timestamp; all
// other fields are user/agent supplied. Optional fields are JSON omitempty so
// the on-disk JSONL stays compact.
type Feedback struct {
	ID               string         `json:"id"`
	Timestamp        time.Time      `json:"timestamp"`
	SessionID        string         `json:"session_id"`
	MessageID        string         `json:"message_id,omitempty"`
	Query            string         `json:"query,omitempty"`
	ResponseType     string         `json:"response_type"`
	Feedback         string         `json:"feedback"`
	Reason           string         `json:"reason,omitempty"`
	PredictedIntent  string         `json:"predicted_intent,omitempty"`
	IntentConfidence float64        `json:"intent_confidence,omitempty"`
	PredictedTicker  string         `json:"predicted_ticker,omitempty"`
	RetrievedDocIDs  []string       `json:"retrieved_doc_ids,omitempty"`
	ResponseExcerpt  string         `json:"response_excerpt,omitempty"`
	Endpoint         string         `json:"endpoint,omitempty"`
	Extra            map[string]any `json:"extra,omitempty"`
}

// Validate returns an error when the record is missing required fields or
// uses an unknown enum value. Sanitize should be called *before* Validate so
// that whitespace-only fields are normalised first.
func (f *Feedback) Validate() error {
	if f == nil {
		return errors.New("feedback: nil record")
	}
	if f.SessionID == "" {
		return errors.New("feedback: session_id is required")
	}
	if !inSet(f.ResponseType, AllowedResponseTypes) {
		return errors.New("feedback: response_type must be one of " + strings.Join(AllowedResponseTypes, ", "))
	}
	if !inSet(f.Feedback, AllowedLabels) {
		return errors.New("feedback: feedback must be one of " + strings.Join(AllowedLabels, ", "))
	}
	return nil
}

// Sanitize trims whitespace, lower-cases enum-like fields, truncates oversize
// strings, and clamps the confidence to [0, 1]. It never returns an error;
// validation is a separate step.
func (f *Feedback) Sanitize() {
	if f == nil {
		return
	}
	f.SessionID = trunc(strings.TrimSpace(f.SessionID), maxSessionIDLen)
	f.MessageID = trunc(strings.TrimSpace(f.MessageID), maxMessageIDLen)
	f.Query = trunc(strings.TrimSpace(f.Query), maxQueryLen)
	f.Reason = trunc(strings.TrimSpace(f.Reason), maxReasonLen)
	f.ResponseExcerpt = trunc(strings.TrimSpace(f.ResponseExcerpt), maxResponseExcerpt)
	f.Endpoint = trunc(strings.TrimSpace(f.Endpoint), maxEndpointLen)
	f.PredictedTicker = strings.ToUpper(strings.TrimSpace(f.PredictedTicker))
	f.PredictedIntent = strings.ToUpper(strings.TrimSpace(f.PredictedIntent))
	f.ResponseType = strings.ToLower(strings.TrimSpace(f.ResponseType))
	f.Feedback = strings.ToLower(strings.TrimSpace(f.Feedback))

	if f.IntentConfidence < 0 {
		f.IntentConfidence = 0
	}
	if f.IntentConfidence > 1 {
		f.IntentConfidence = 1
	}

	if len(f.RetrievedDocIDs) > maxDocIDs {
		f.RetrievedDocIDs = f.RetrievedDocIDs[:maxDocIDs]
	}
	for i, d := range f.RetrievedDocIDs {
		f.RetrievedDocIDs[i] = trunc(strings.TrimSpace(d), maxDocIDLen)
	}
}

// ListFilter narrows down which records List returns. Empty fields mean
// "no filter on that dimension".
type ListFilter struct {
	Feedback     string
	ResponseType string
	SessionID    string
	Since        time.Time
	Limit        int
}

func inSet(s string, allowed []string) bool {
	for _, a := range allowed {
		if s == a {
			return true
		}
	}
	return false
}

func trunc(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

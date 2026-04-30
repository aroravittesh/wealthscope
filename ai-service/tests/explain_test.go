package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wealthscope-ai/internal/explain"
	"wealthscope-ai/internal/finsentiment"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/newsentiment"
	"wealthscope-ai/internal/portfoliorisk"
	"wealthscope-ai/internal/prediction"
	"wealthscope-ai/internal/rag"
)

// --- Intent: keyword path produces matched keywords + summary --------------

func TestExplain_Intent_KeywordMatchHasSignals(t *testing.T) {
	r := ml.DetectIntentKeywords("What is the current price of $AAPL?")
	if r.Intent != ml.IntentStockPrice {
		t.Fatalf("intent: want STOCK_PRICE got %s", r.Intent)
	}
	if r.Source != ml.IntentSourceKeyword {
		t.Fatalf("source: want %q got %q", ml.IntentSourceKeyword, r.Source)
	}
	if len(r.MatchedKeywords) == 0 {
		t.Fatal("expected matched keywords for keyword classification")
	}
	if r.Explanation == nil {
		t.Fatal("explanation must be populated")
	}
	if r.Explanation.Code != "INTENT_KEYWORD_MATCH" {
		t.Fatalf("code: want INTENT_KEYWORD_MATCH got %s", r.Explanation.Code)
	}
	if !strings.Contains(strings.ToLower(r.Explanation.Summary), "stock_price") {
		t.Fatalf("summary should mention intent: %q", r.Explanation.Summary)
	}
	if !hasSignalCode(r.Explanation.TopSignals, "ENTITY_TICKER") {
		t.Fatalf("expected ENTITY_TICKER signal, got %v", signalCodes(r.Explanation.TopSignals))
	}
	if !hasSignalCode(r.Explanation.TopSignals, "INTENT_KEYWORD") {
		t.Fatalf("expected INTENT_KEYWORD signal, got %v", signalCodes(r.Explanation.TopSignals))
	}
}

// --- Intent: remote path explains as remote_classifier --------------------

func TestExplain_Intent_RemoteHighConfidence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"intent":"MARKET_NEWS","confidence":0.92}`))
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "what's happening with markets today")
	if r.Source != ml.IntentSourceRemote {
		t.Fatalf("source: want %q got %q", ml.IntentSourceRemote, r.Source)
	}
	if r.Explanation == nil || r.Explanation.Code != "INTENT_REMOTE_HIGH_CONFIDENCE" {
		t.Fatalf("expected INTENT_REMOTE_HIGH_CONFIDENCE explanation, got %+v", r.Explanation)
	}
	if !hasSignalCode(r.Explanation.TopSignals, "INTENT_REMOTE") {
		t.Fatalf("expected INTENT_REMOTE signal, got %v", signalCodes(r.Explanation.TopSignals))
	}
}

// --- Intent: low-confidence remote falls back AND explanation says so ------

func TestExplain_Intent_LowConfidenceFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"intent":"GENERAL_MARKET","confidence":0.20}`))
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client(), MinConfidence: 0.5}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "What is the current price of AAPL?")
	if r.Source != ml.IntentSourceLowConfFallback {
		t.Fatalf("source: want %q got %q", ml.IntentSourceLowConfFallback, r.Source)
	}
	if r.Explanation == nil || r.Explanation.Code != "INTENT_REMOTE_LOW_CONFIDENCE_FALLBACK" {
		t.Fatalf("expected low-conf fallback code, got %+v", r.Explanation)
	}
	// Reasons should include the original remote prediction prefix prepended in finalize path.
	joined := strings.Join(r.Explanation.Reasons, " | ")
	if !strings.Contains(joined, "Remote classifier returned GENERAL_MARKET") {
		t.Fatalf("reasons should mention what remote said: %q", joined)
	}
}

// --- Intent: remote error → explanation reflects fallback path -------------

func TestExplain_Intent_RemoteErrorFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "How should I diversify my portfolio?")
	if r.Source != ml.IntentSourceRemoteFallback {
		t.Fatalf("source: want %q got %q", ml.IntentSourceRemoteFallback, r.Source)
	}
	if r.Explanation == nil || r.Explanation.Code != "INTENT_REMOTE_ERROR_FALLBACK" {
		t.Fatalf("expected error-fallback code, got %+v", r.Explanation)
	}
}

// --- Sentiment: bullish article produces positive signals + reason code ----

func TestExplain_Sentiment_BullishHasPositiveSignals(t *testing.T) {
	r := newsentiment.Aggregate("TSLA", []market.NewsItem{
		{Title: "Stock surges on strong earnings beat and bullish guidance", Description: "Rally continues with record profit and upbeat outlook"},
	})
	if r.OverallSentiment != string(ml.SentimentBullish) {
		t.Fatalf("overall: want BULLISH got %s", r.OverallSentiment)
	}
	if len(r.TopPositiveSignals) == 0 {
		t.Fatal("expected positive signals")
	}
	if r.ReasonCode != "SENT_BULLISH_DOMINANT_POSITIVE" {
		t.Fatalf("reason_code: want SENT_BULLISH_DOMINANT_POSITIVE got %s", r.ReasonCode)
	}
	if r.ExplanationDetail == nil || !strings.Contains(strings.ToLower(r.ReasoningSummary), "bullish") {
		t.Fatalf("reasoning should mention bullish: %q", r.ReasoningSummary)
	}
}

// --- Sentiment: empty articles still emit a structured explanation ---------

func TestExplain_Sentiment_EmptyArticlesExplains(t *testing.T) {
	r := newsentiment.Aggregate("AAPL", nil)
	if r.ReasonCode != "SENT_NO_ARTICLES" {
		t.Fatalf("expected SENT_NO_ARTICLES, got %s", r.ReasonCode)
	}
	if r.ExplanationDetail == nil || r.ExplanationDetail.Disclaimer == "" {
		t.Fatal("empty-articles explanation should still carry a disclaimer")
	}
}

// --- Sentiment builder: pos/neg signals split correctly -------------------

func TestExplain_Sentiment_BuilderSplitsSignals(t *testing.T) {
	pos, neg := explain.SignalsFromTerms([]finsentiment.TermHit{
		{Term: "earnings beat", Polarity: 0.8},
		{Term: "downgrade", Polarity: -0.7},
	})
	if len(pos) != 1 || pos[0].Code != "FIN_PHRASE_BULLISH" {
		t.Fatalf("positive signals wrong: %+v", pos)
	}
	if len(neg) != 1 || neg[0].Code != "FIN_TERM_BEARISH" {
		t.Fatalf("negative signals wrong: %+v", neg)
	}
}

// --- Risk: composite risk explanation reflects top driver ------------------

func TestExplain_Risk_DriversReflectAssessment(t *testing.T) {
	report := ml.ScorePortfolio([]ml.PortfolioHolding{
		{Symbol: "TSLA", Allocation: 0.7, Beta: "2.0"},
		{Symbol: "NVDA", Allocation: 0.3, Beta: "1.8"},
	})
	if report.ExplanationDetail == nil {
		t.Fatal("risk explanation should be populated")
	}
	if report.ReasonCode == "" || !strings.HasPrefix(report.ReasonCode, "RISK_") {
		t.Fatalf("reason code should be RISK_*, got %q", report.ReasonCode)
	}
	if len(report.ExplanationDetail.TopSignals) == 0 {
		t.Fatal("expected top signals on risk explanation")
	}
	if !hasSignalPrefix(report.ExplanationDetail.TopSignals, "RISK_DRIVER_") {
		t.Fatalf("expected RISK_DRIVER_* signals, got %v", signalCodes(report.ExplanationDetail.TopSignals))
	}
}

// --- Drift: explanation mentions target + populates envelope ---------------

func TestExplain_Drift_TargetMisalignmentInExplanation(t *testing.T) {
	resp, err := prediction.PredictRiskDrift(prediction.DriftRequest{
		TargetRisk: "LOW",
		Holdings: []prediction.DriftHolding{
			{Symbol: "TSLA", Allocation: 0.7, Beta: "2.0", Sector: "Technology"},
			{Symbol: "NVDA", Allocation: 0.3, Beta: "1.8", Sector: "Technology"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExplanationDetail == nil {
		t.Fatal("drift explanation_detail should be populated")
	}
	if resp.ReasonCode == "" || !strings.Contains(resp.ReasonCode, "DRIFT") {
		t.Fatalf("reason code should mention DRIFT, got %q", resp.ReasonCode)
	}
	low := strings.ToLower(resp.ReasoningSummary)
	if !strings.Contains(low, "low") {
		t.Fatalf("drift summary should mention LOW target, got %q", resp.ReasoningSummary)
	}
	if !hasSignalPrefix(resp.ExplanationDetail.TopSignals, "DRIFT_DRIVER_") {
		t.Fatalf("expected DRIFT_DRIVER_* signals, got %v", signalCodes(resp.ExplanationDetail.TopSignals))
	}
	// Legacy field still populated for back-compat.
	if resp.Explanation == "" {
		t.Fatal("legacy explanation string must remain populated")
	}
}

// --- Risk explanation builder: handles empty drivers gracefully ------------

func TestExplain_Risk_BuilderHandlesEmptyDrivers(t *testing.T) {
	exp := explain.BuildRiskExplanation(explain.RiskInputs{
		Level:   "LOW",
		Score:   0.05,
		Drivers: nil,
	})
	if exp.Code != "RISK_LOW_GENERIC" {
		t.Fatalf("code: want RISK_LOW_GENERIC got %s", exp.Code)
	}
	if !strings.Contains(strings.ToLower(exp.Summary), "low") {
		t.Fatalf("summary should mention low: %q", exp.Summary)
	}
}

// --- Drift explanation builder: includes weighted-beta vs reference --------

func TestExplain_Drift_BuilderIncludesBetaGap(t *testing.T) {
	exp := explain.BuildDriftExplanation(explain.DriftInputs{
		Level:        "HIGH_DRIFT",
		Score:        0.7,
		Target:       "LOW",
		WeightedBeta: 1.9,
		CenterBeta:   0.65,
		Misalignment: 1.0,
		Drivers: []portfoliorisk.Driver{
			{Code: "MISALIGNMENT", Label: "Distance from risk target", Contribution: 0.45, Value: 1.0},
		},
	})
	if exp.Code != "HIGH_DRIFT_DRIVEN_BY_MISALIGNMENT" {
		t.Fatalf("code: want HIGH_DRIFT_DRIVEN_BY_MISALIGNMENT got %s", exp.Code)
	}
	joined := strings.Join(exp.Reasons, " ")
	if !strings.Contains(joined, "1.90") || !strings.Contains(joined, "0.65") {
		t.Fatalf("reasons should include beta numbers: %q", joined)
	}
}

// --- Retrieval: hits include sub-scores + structured explanation -----------

func TestExplain_Retrieval_HitsHaveSubScoresAndExplanation(t *testing.T) {
	hits := rag.RetrieveWithExplanation("dividend yield income investors", rag.RetrievalContext{}, 3)
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
	first := hits[0]
	if first.Reason == "" {
		t.Fatal("expected ranking reason on first hit")
	}
	if first.Explanation.Code != "RAG_HYBRID_RANKING" {
		t.Fatalf("expected RAG_HYBRID_RANKING code, got %q", first.Explanation.Code)
	}
	codes := signalCodes(first.Explanation.TopSignals)
	for _, want := range []string{"RAG_SEMANTIC", "RAG_LEXICAL", "RAG_ENTITY", "RAG_METADATA"} {
		if !contains(codes, want) {
			t.Fatalf("expected signal %s in retrieval explanation, got %v", want, codes)
		}
	}
}

// --- Retrieval: no-match query stays empty --------------------------------

func TestExplain_Retrieval_NoMatchEmpty(t *testing.T) {
	hits := rag.RetrieveWithExplanation("xyzzy abracadabra grommet", rag.RetrievalContext{}, 3)
	if len(hits) != 0 {
		t.Fatalf("expected empty hits, got %d", len(hits))
	}
}

// --- Helpers --------------------------------------------------------------

func hasSignalCode(signals []explain.Signal, code string) bool {
	for _, s := range signals {
		if s.Code == code {
			return true
		}
	}
	return false
}

func hasSignalPrefix(signals []explain.Signal, prefix string) bool {
	for _, s := range signals {
		if strings.HasPrefix(s.Code, prefix) {
			return true
		}
	}
	return false
}

func signalCodes(signals []explain.Signal) []string {
	out := make([]string, len(signals))
	for i, s := range signals {
		out[i] = s.Code
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

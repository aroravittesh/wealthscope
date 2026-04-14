package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"wealthscope-ai/internal/ml"
)

func TestDetectIntent_RemotePrimary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/classify-intent" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"intent":"MARKET_NEWS","confidence":0.91}`))
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "dummy message about headlines")
	if r.Intent != ml.IntentMarketNews {
		t.Fatalf("intent: want MARKET_NEWS got %s", r.Intent)
	}
	if r.Confidence != 0.91 {
		t.Fatalf("confidence: want 0.91 got %f", r.Confidence)
	}
}

func TestDetectIntent_RemoteFallbackOnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "What is the current price of AAPL?")
	if r.Intent != ml.IntentStockPrice {
		t.Fatalf("fallback intent: want STOCK_PRICE got %s", r.Intent)
	}
	if r.Ticker != "AAPL" {
		t.Fatalf("ticker: want AAPL got %s", r.Ticker)
	}
}

func TestDetectIntent_RemoteFallbackInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "How should I diversify my portfolio?")
	if r.Intent != ml.IntentPortfolioTip {
		t.Fatalf("fallback intent: want PORTFOLIO_TIP got %s", r.Intent)
	}
}

func TestDetectIntent_RemoteFallbackUnknownLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"intent":"BUY_PIZZA","confidence":0.99}`))
	}))
	defer srv.Close()

	cfg := ml.IntentConfig{ClassifierBaseURL: srv.URL, Client: srv.Client()}
	r := ml.DetectIntentWithConfig(context.Background(), cfg, "What is the latest news on MSFT?")
	if r.Intent != ml.IntentMarketNews {
		t.Fatalf("fallback intent: want MARKET_NEWS got %s", r.Intent)
	}
}

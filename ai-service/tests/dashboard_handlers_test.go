package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/handler"
	"wealthscope-ai/internal/ml"
)

func dashboardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/portfolio/explain", handler.PortfolioExplainHandler)
	r.POST("/portfolio/summarize", handler.PortfolioSummarizeHandler)
	r.POST("/portfolio/changes", handler.PortfolioChangesHandler)
	r.POST("/compare", handler.CompareHandler)
	r.GET("/news-sentiment/:symbol", handler.NewsSentimentHandler)
	return r
}

func TestHTTP_PortfolioSummarize_OK(t *testing.T) {
	body := map[string]any{
		"holdings": []map[string]any{
			{"symbol": "A", "allocation": 0.5, "beta": "1"},
			{"symbol": "B", "allocation": 0.5, "beta": "1"},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/summarize", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if _, ok := out["summary"]; !ok {
		t.Fatalf("missing summary: %v", out)
	}
}

func TestHTTP_PortfolioSummarize_BadJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/summarize", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHTTP_PortfolioChanges_NoPrior(t *testing.T) {
	body := map[string]any{
		"current": map[string]any{
			"holdings": []map[string]any{
				{"symbol": "TSLA", "allocation": 1, "beta": "1.5"},
			},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/changes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var out struct {
		HasPriorSnapshot bool `json:"has_prior_snapshot"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.HasPriorSnapshot {
		t.Fatal("expected false")
	}
}

func TestHTTP_PortfolioChanges_WithPrior(t *testing.T) {
	body := map[string]any{
		"current": map[string]any{
			"holdings": []map[string]any{
				{"symbol": "AAPL", "allocation": 0.5, "beta": "1"},
				{"symbol": "MSFT", "allocation": 0.5, "beta": "1"},
			},
		},
		"prior": map[string]any{
			"holdings": []map[string]any{
				{"symbol": "AAPL", "allocation": 1, "beta": "1"},
			},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/changes", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var out struct {
		HasPriorSnapshot bool `json:"has_prior_snapshot"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	if !out.HasPriorSnapshot {
		t.Fatal("expected prior comparison")
	}
}

func TestHTTP_PortfolioExplain_OK(t *testing.T) {
	body := map[string]any{
		"holdings": []ml.PortfolioHolding{
			{Symbol: "AAPL", Allocation: 1, Beta: "1.1"},
		},
		"target_risk": "MEDIUM",
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/explain", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestHTTP_Compare_InvalidCount(t *testing.T) {
	body := map[string]any{"symbols": []string{"AAPL"}}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestHTTP_NewsSentiment_BadGatewayOrOK(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/news-sentiment/AAPL", nil)
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status %d %s", w.Code, w.Body.String())
	}
}

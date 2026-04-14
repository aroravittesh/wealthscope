package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/handler"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/prediction"
)

func dashboardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/portfolio/explain", handler.PortfolioExplainHandler)
	r.POST("/portfolio/summarize", handler.PortfolioSummarizeHandler)
	r.POST("/portfolio/changes", handler.PortfolioChangesHandler)
	r.POST("/compare", handler.CompareHandler)
	r.POST("/predict/risk-drift", handler.RiskDriftHandler)
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

func TestHTTP_Compare_OK_WithMockFetcher(t *testing.T) {
	f := mockFetcher{
		quotes: map[string]*market.GlobalQuote{
			"AAPL": {Symbol: "AAPL", Price: "180"},
			"MSFT": {Symbol: "MSFT", Price: "400"},
		},
		overviews: map[string]*market.CompanyOverview{
			"AAPL": {Sector: "Technology", Beta: "1.1", PERatio: "22", MarketCap: "1000"},
			"MSFT": {Sector: "Technology", Beta: "1.0", PERatio: "30", MarketCap: "2000"},
		},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/compare", handler.CompareHandlerWithFetcher(f))

	body := map[string]any{"symbols": []string{"AAPL", "MSFT"}}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Comparisons []struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		} `json:"comparisons"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Comparisons) != 2 || out.Summary == "" {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestHTTP_Compare_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/compare", handler.CompareHandlerWithFetcher(mockFetcher{}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHTTP_Compare_UpstreamQuoteError_Is502(t *testing.T) {
	f := mockFetcher{
		quotes: map[string]*market.GlobalQuote{"MSFT": {Price: "1"}},
		quoteErr: map[string]error{
			"AAPL": errors.New("upstream unavailable"),
		},
		overviews: map[string]*market.CompanyOverview{
			"AAPL": {Beta: "1"},
			"MSFT": {Beta: "1"},
		},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/compare", handler.CompareHandlerWithFetcher(f))
	body := map[string]any{"symbols": []string{"AAPL", "MSFT"}}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 got %d body %s", w.Code, w.Body.String())
	}
}

type staticNewsFetcher struct {
	items []market.NewsItem
	err   error
}

func (s staticNewsFetcher) FetchNews(symbol string) ([]market.NewsItem, error) {
	return s.items, s.err
}

func TestHTTP_NewsSentiment_OK_WithMockFetcher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	f := staticNewsFetcher{
		items: []market.NewsItem{
			{Title: "Stock gains on strong outlook", Description: "Rally continues"},
		},
	}
	r.GET("/news-sentiment/:symbol", handler.NewsSentimentHandlerWithFetcher(f))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/news-sentiment/tsla", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Symbol           string  `json:"symbol"`
		OverallSentiment string  `json:"overall_sentiment"`
		ArticleCount     int     `json:"article_count"`
		Confidence       float64 `json:"confidence"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Symbol != "TSLA" || out.ArticleCount != 1 {
		t.Fatalf("unexpected %+v", out)
	}
}

func TestHTTP_NewsSentiment_FetchError_Is502(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/news-sentiment/:symbol", handler.NewsSentimentHandlerWithFetcher(staticNewsFetcher{err: errors.New("network down")}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/news-sentiment/AAPL", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 got %d", w.Code)
	}
}

func TestHTTP_PortfolioExplain_BadJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/explain", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHTTP_PortfolioExplain_EmptyHoldings(t *testing.T) {
	body := map[string]any{
		"holdings":    []ml.PortfolioHolding{},
		"target_risk": "MEDIUM",
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/portfolio/explain", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d %s", w.Code, w.Body.String())
	}
}

func TestHTTP_RiskDrift_OK(t *testing.T) {
	body := map[string]any{
		"target_risk": "MEDIUM",
		"holdings": []map[string]any{
			{"symbol": "A", "allocation": 0.5, "beta": "1.0", "sector": "X"},
			{"symbol": "B", "allocation": 0.5, "beta": "1.0", "sector": "Y"},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/predict/risk-drift", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var out prediction.DriftResponse
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.DriftLevel == "" {
		t.Fatal("expected drift_level")
	}
}

func TestHTTP_RiskDrift_EmptyHoldings(t *testing.T) {
	body := map[string]any{"target_risk": "LOW", "holdings": []any{}}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/predict/risk-drift", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}

func TestHTTP_RiskDrift_BadJSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/predict/risk-drift", bytes.NewBufferString("x"))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHTTP_RiskDrift_InvalidTarget(t *testing.T) {
	body := map[string]any{
		"target_risk": "BANANA",
		"holdings": []map[string]any{
			{"symbol": "A", "allocation": 1, "beta": "1", "sector": "S"},
		},
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/predict/risk-drift", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	dashboardRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d %s", w.Code, w.Body.String())
	}
}

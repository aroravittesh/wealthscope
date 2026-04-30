package intentremote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const classifyPath = "/classify-intent"

// Result is a successful response from the intent classifier service.
type Result struct {
	Intent     string
	Confidence float64
}

type classifyRequest struct {
	Message string `json:"message"`
}

type classifyResponse struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
}

// Classify calls the Python inference API. ok is false on network, HTTP, or JSON errors.
func Classify(ctx context.Context, client *http.Client, baseURL, message string) (Result, bool) {
	baseURL = strings.TrimSpace(strings.TrimSuffix(baseURL, "/"))
	if baseURL == "" || message == "" {
		return Result{}, false
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	body, err := json.Marshal(classifyRequest{Message: message})
	if err != nil {
		return Result{}, false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+classifyPath, bytes.NewReader(body))
	if err != nil {
		return Result{}, false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, false
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, false
	}
	if resp.StatusCode != http.StatusOK {
		return Result{}, false
	}

	var parsed classifyResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Result{}, false
	}
	if parsed.Intent == "" || parsed.Confidence < 0 || parsed.Confidence > 1 {
		return Result{}, false
	}
	if err := validateIntentLabel(parsed.Intent); err != nil {
		return Result{}, false
	}

	return Result{Intent: parsed.Intent, Confidence: parsed.Confidence}, true
}

func validateIntentLabel(s string) error {
	switch s {
	case "STOCK_PRICE", "RISK_ANALYSIS", "MARKET_NEWS", "PORTFOLIO_TIP", "GENERAL_MARKET", "UNKNOWN":
		return nil
	default:
		return fmt.Errorf("unknown intent label %q", s)
	}
}

package websearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultTavilyEndpoint is the upstream Tavily search endpoint. Override in
// tests via NewTavilyProviderWithURL.
const DefaultTavilyEndpoint = "https://api.tavily.com/search"

// TavilyProvider calls the Tavily search API. Free / paid keys both work; the
// "topic":"news" parameter biases results toward recent articles.
type TavilyProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewTavilyProvider builds a provider with the default endpoint.
func NewTavilyProvider(apiKey string, client *http.Client) *TavilyProvider {
	return NewTavilyProviderWithURL(apiKey, DefaultTavilyEndpoint, client)
}

// NewTavilyProviderWithURL allows overriding the endpoint (tests, mock
// servers, future enterprise proxies).
func NewTavilyProviderWithURL(apiKey, endpoint string, client *http.Client) *TavilyProvider {
	if client == nil {
		client = http.DefaultClient
	}
	return &TavilyProvider{
		apiKey:   strings.TrimSpace(apiKey),
		endpoint: endpoint,
		client:   client,
	}
}

// Name returns "tavily".
func (TavilyProvider) Name() string { return "tavily" }

type tavilyRequest struct {
	APIKey            string `json:"api_key"`
	Query             string `json:"query"`
	Topic             string `json:"topic"`
	SearchDepth       string `json:"search_depth"`
	MaxResults        int    `json:"max_results"`
	IncludeAnswer     bool   `json:"include_answer"`
	IncludeRawContent bool   `json:"include_raw_content"`
	Days              int    `json:"days,omitempty"`
}

type tavilyResult struct {
	Title         string  `json:"title"`
	URL           string  `json:"url"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	PublishedDate string  `json:"published_date"`
}

type tavilyResponse struct {
	Results []tavilyResult `json:"results"`
}

// Search calls Tavily and returns up to maxResults raw hits. Errors at the
// HTTP / JSON layer are returned as-is so the caller can decide whether to
// degrade silently (the chat pipeline does).
func (p *TavilyProvider) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	if p.apiKey == "" {
		return nil, errors.New("tavily: missing API key")
	}
	if maxResults <= 0 || maxResults > 10 {
		maxResults = 5
	}

	body := tavilyRequest{
		APIKey:        p.apiKey,
		Query:         strings.TrimSpace(query),
		Topic:         "news",
		SearchDepth:   "basic",
		MaxResults:    maxResults,
		IncludeAnswer: false,
		Days:          7,
	}
	enc, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tavily: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(enc))
	if err != nil {
		return nil, fmt.Errorf("tavily: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("tavily: status %d: %s", resp.StatusCode, string(raw))
	}

	var parsed tavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("tavily: decode: %w", err)
	}

	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{
			Title:       r.Title,
			Snippet:     r.Content,
			URL:         r.URL,
			Source:      hostnameOf(r.URL, ""),
			PublishedAt: r.PublishedDate,
			Score:       r.Score,
		})
	}
	return out, nil
}

package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PriceProvider resolves an estimated unit price (per share/coin) for analytics.
// avgPrice is the holding's average cost; implementations may use it as fallback.
type PriceProvider interface {
	UnitPrice(symbol string, avgPrice float64) float64
}

// Passthrough uses average cost as the market price (no external data).
type Passthrough struct{}

func (Passthrough) UnitPrice(_ string, avgPrice float64) float64 {
	return avgPrice
}

// Simulated applies a deterministic multiplier in [0.92, 1.08] from the symbol
// (similar range to the dashboard demo) so P/L is non-zero without an API key.
type Simulated struct{}

func (Simulated) UnitPrice(symbol string, avgPrice float64) float64 {
	if avgPrice <= 0 {
		return 0
	}
	return avgPrice * DeterministicMultiplier(symbol)
}

// DeterministicMultiplier returns a stable value in [0.92, 1.08] for a symbol.
func DeterministicMultiplier(symbol string) float64 {
	h := uint32(5381)
	for i := 0; i < len(symbol); i++ {
		h = h*33 + uint32(symbol[i])
	}
	return 0.92 + (float64(h%10000)/10000.0)*0.16
}

type quoteResponse struct {
	GlobalQuote struct {
		Symbol string `json:"01. symbol"`
		Price  string `json:"05. price"`
	} `json:"Global Quote"`
}

type alphaProvider struct {
	apiKey string
	client *http.Client
	ttl    time.Duration

	mu    sync.Mutex
	cache map[string]cachedQuote
}

type cachedQuote struct {
	price float64
	at    time.Time
}

func newAlphaProvider(apiKey string) *alphaProvider {
	return &alphaProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
		ttl:    55 * time.Second,
		cache:  make(map[string]cachedQuote),
	}
}

func (a *alphaProvider) UnitPrice(symbol string, avgPrice float64) float64 {
	sym := strings.TrimSpace(strings.ToUpper(symbol))
	if sym == "" {
		return avgPrice
	}

	a.mu.Lock()
	if ent, ok := a.cache[sym]; ok && time.Since(ent.at) < a.ttl && ent.price > 0 {
		p := ent.price
		a.mu.Unlock()
		return p
	}
	a.mu.Unlock()

	p, err := a.fetchQuote(sym)
	if err != nil || p <= 0 {
		return avgPrice
	}

	a.mu.Lock()
	a.cache[sym] = cachedQuote{price: p, at: time.Now()}
	a.mu.Unlock()
	return p
}

func (a *alphaProvider) fetchQuote(symbol string) (float64, error) {
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		symbol, a.apiKey,
	)
	resp, err := a.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, err
	}

	var q quoteResponse
	if err := json.Unmarshal(body, &q); err != nil {
		return 0, err
	}
	priceStr := strings.TrimSpace(q.GlobalQuote.Price)
	if priceStr == "" {
		return 0, fmt.Errorf("no price in quote")
	}
	return strconv.ParseFloat(priceStr, 64)
}

// NewDefaultProvider uses Alpha Vantage when ALPHA_VANTAGE_API_KEY is set; otherwise Simulated.
func NewDefaultProvider() PriceProvider {
	if k := strings.TrimSpace(os.Getenv("ALPHA_VANTAGE_API_KEY")); k != "" {
		return newAlphaProvider(k)
	}
	return Simulated{}
}

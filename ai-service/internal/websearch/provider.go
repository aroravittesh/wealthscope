package websearch

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Environment variable names. WEALTHSCOPE_WEB_SEARCH_PROVIDER explicitly picks
// a provider ("tavily" | "stub" | "off"); when unset we autodetect from the
// presence of provider-specific API keys.
const (
	EnvProvider     = "WEALTHSCOPE_WEB_SEARCH_PROVIDER"
	EnvTavilyAPIKey = "TAVILY_API_KEY"
)

// StubProvider is the no-op provider used when web search is disabled or
// unconfigured. It always returns an empty result set with no error so the
// chat pipeline degrades gracefully.
type StubProvider struct{}

// Name returns the canonical provider name ("stub").
func (StubProvider) Name() string { return "stub" }

// Search returns no results and no error.
func (StubProvider) Search(_ context.Context, _ string, _ int) ([]Result, error) {
	return nil, nil
}

// MockProvider is a deterministic in-memory provider used in tests. The Calls
// field records every Search invocation for assertions.
type MockProvider struct {
	NameStr string
	Results []Result
	Err     error

	mu    sync.Mutex
	Calls []MockCall
}

// MockCall captures the arguments passed to MockProvider.Search.
type MockCall struct {
	Query      string
	MaxResults int
}

// Name returns the configured mock name (default "mock").
func (m *MockProvider) Name() string {
	if strings.TrimSpace(m.NameStr) == "" {
		return "mock"
	}
	return m.NameStr
}

// Search returns the configured Results / Err, recording the call.
func (m *MockProvider) Search(_ context.Context, q string, k int) ([]Result, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, MockCall{Query: q, MaxResults: k})
	m.mu.Unlock()
	if m.Err != nil {
		return nil, m.Err
	}
	out := make([]Result, len(m.Results))
	copy(out, m.Results)
	return out, nil
}

// ---- process-wide default provider ----

var (
	defaultProviderOnce sync.Once
	defaultProviderMu   sync.RWMutex
	defaultProvider     Provider
)

// DefaultProvider returns the process-wide provider, lazily initialising it
// from environment on first use:
//
//   - WEALTHSCOPE_WEB_SEARCH_PROVIDER=off  → StubProvider
//   - WEALTHSCOPE_WEB_SEARCH_PROVIDER=stub → StubProvider
//   - WEALTHSCOPE_WEB_SEARCH_PROVIDER=tavily and TAVILY_API_KEY set → TavilyProvider
//   - Unset and TAVILY_API_KEY set → TavilyProvider
//   - Otherwise → StubProvider
func DefaultProvider() Provider {
	defaultProviderOnce.Do(func() {
		defaultProviderMu.Lock()
		defer defaultProviderMu.Unlock()
		if defaultProvider != nil {
			return
		}
		defaultProvider = providerFromEnv()
	})
	defaultProviderMu.RLock()
	defer defaultProviderMu.RUnlock()
	return defaultProvider
}

// SetDefaultProviderForTest overrides the process-wide provider. Returns a
// cleanup that restores the previous one.
func SetDefaultProviderForTest(p Provider) (cleanup func()) {
	defaultProviderMu.Lock()
	prev := defaultProvider
	defaultProvider = p
	defaultProviderMu.Unlock()
	// Mark as initialised so DefaultProvider() does not overwrite this if
	// another goroutine races to first-call it.
	defaultProviderOnce.Do(func() {})
	return func() {
		defaultProviderMu.Lock()
		defaultProvider = prev
		defaultProviderMu.Unlock()
	}
}

func providerFromEnv() Provider {
	choice := strings.ToLower(strings.TrimSpace(os.Getenv(EnvProvider)))
	tavilyKey := strings.TrimSpace(os.Getenv(EnvTavilyAPIKey))

	switch choice {
	case "off", "stub", "none", "disabled":
		return StubProvider{}
	case "tavily":
		if tavilyKey == "" {
			return StubProvider{}
		}
		return NewTavilyProvider(tavilyKey, &http.Client{Timeout: 4 * time.Second})
	}

	// Auto-detect.
	if tavilyKey != "" {
		return NewTavilyProvider(tavilyKey, &http.Client{Timeout: 4 * time.Second})
	}
	return StubProvider{}
}

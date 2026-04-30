package websearch

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
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
// from config wiring on first use. If unset, StubProvider is used.
func DefaultProvider() Provider {
	defaultProviderOnce.Do(func() {
		defaultProviderMu.Lock()
		defer defaultProviderMu.Unlock()
		if defaultProvider != nil {
			return
		}
		defaultProvider = StubProvider{}
	})
	defaultProviderMu.RLock()
	defer defaultProviderMu.RUnlock()
	return defaultProvider
}

type ProviderConfig struct {
	Provider  string
	TavilyKey string
	Timeout   time.Duration
}

// SetDefaultProviderFromConfig wires the provider based on startup config.
// Supported providers: tavily, stub/off/none/disabled.
func SetDefaultProviderFromConfig(cfg ProviderConfig) {
	defaultProviderMu.Lock()
	defer defaultProviderMu.Unlock()
	defaultProvider = providerFromConfig(cfg)
	defaultProviderOnce.Do(func() {})
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

func providerFromConfig(cfg ProviderConfig) Provider {
	choice := strings.ToLower(strings.TrimSpace(cfg.Provider))
	tavilyKey := strings.TrimSpace(cfg.TavilyKey)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	switch choice {
	case "off", "stub", "none", "disabled":
		return StubProvider{}
	case "tavily":
		if tavilyKey == "" {
			return StubProvider{}
		}
		return NewTavilyProvider(tavilyKey, &http.Client{Timeout: timeout})
	}

	// Auto-detect when provider omitted.
	if tavilyKey != "" {
		return NewTavilyProvider(tavilyKey, &http.Client{Timeout: timeout})
	}
	return StubProvider{}
}

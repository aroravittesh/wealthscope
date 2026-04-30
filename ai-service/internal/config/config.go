package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	Port string
}

type OpenAIConfig struct {
	APIKey string
}

type MarketConfig struct {
	AlphaVantageAPIKey string
	NewsAPIKey         string
}

type IntentConfig struct {
	ClassifierURL string
	MinConfidence float64
	Timeout       time.Duration
}

type WebSearchConfig struct {
	Provider     string
	TavilyAPIKey string
	Enabled      bool
	Timeout      time.Duration
}

type SessionConfig struct {
	TTL              time.Duration
	MaxMessages      int
	KeepAfterCompact int
}

type FeedbackConfig struct {
	Path string
}

type RAGConfig struct {
	QADatasetPath string
}

type AppConfig struct {
	Server    ServerConfig
	OpenAI    OpenAIConfig
	Market    MarketConfig
	Intent    IntentConfig
	WebSearch WebSearchConfig
	Session   SessionConfig
	Feedback  FeedbackConfig
	RAG       RAGConfig
}

func LoadFromEnv() (AppConfig, error) {
	cfg := AppConfig{
		Server: ServerConfig{
			Port: envOrDefault("PORT", "9000"),
		},
		OpenAI: OpenAIConfig{
			APIKey: strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		},
		Market: MarketConfig{
			AlphaVantageAPIKey: strings.TrimSpace(os.Getenv("ALPHA_VANTAGE_API_KEY")),
			NewsAPIKey:         strings.TrimSpace(os.Getenv("NEWS_API_KEY")),
		},
		Intent: IntentConfig{
			ClassifierURL: strings.TrimSuffix(strings.TrimSpace(os.Getenv("INTENT_CLASSIFIER_URL")), "/"),
			MinConfidence: parseFloat01(os.Getenv("INTENT_MIN_CONFIDENCE"), 0),
			Timeout:       parseDurationOrDefault(os.Getenv("WEALTHSCOPE_INTENT_TIMEOUT"), 5*time.Second),
		},
		WebSearch: WebSearchConfig{
			Provider:     strings.ToLower(strings.TrimSpace(os.Getenv("WEALTHSCOPE_WEB_SEARCH_PROVIDER"))),
			TavilyAPIKey: strings.TrimSpace(os.Getenv("TAVILY_API_KEY")),
			Timeout:      parseDurationOrDefault(os.Getenv("WEALTHSCOPE_WEBSEARCH_TIMEOUT"), 4*time.Second),
		},
		Session: SessionConfig{
			TTL:              parseDurationOrDefault(os.Getenv("WEALTHSCOPE_SESSION_TTL"), 24*time.Hour),
			MaxMessages:      parseIntOrDefault(os.Getenv("WEALTHSCOPE_SESSION_MAX_MESSAGES"), 24),
			KeepAfterCompact: parseIntOrDefault(os.Getenv("WEALTHSCOPE_SESSION_KEEP_AFTER_COMPACT"), 10),
		},
		Feedback: FeedbackConfig{
			Path: envOrDefault("WEALTHSCOPE_FEEDBACK_PATH", "data/feedback.jsonl"),
		},
		RAG: RAGConfig{
			QADatasetPath: strings.TrimSpace(os.Getenv("WEALTHSCOPE_QA_DATASET_PATH")),
		},
	}

	cfg.WebSearch.Enabled = webSearchEnabled(cfg.WebSearch.Provider)

	if cfg.Server.Port == "" {
		return AppConfig{}, fmt.Errorf("config: PORT is empty")
	}
	return cfg, nil
}

func (c AppConfig) SafeSummary() string {
	return fmt.Sprintf(
		"port=%s intent_url_set=%t websearch_enabled=%t websearch_provider=%s qa_dataset_override=%t feedback_path=%s session_ttl=%s",
		c.Server.Port,
		c.Intent.ClassifierURL != "",
		c.WebSearch.Enabled,
		orDash(c.WebSearch.Provider),
		c.RAG.QADatasetPath != "",
		c.Feedback.Path,
		c.Session.TTL.String(),
	)
}

func parseDurationOrDefault(raw string, def time.Duration) time.Duration {
	s := strings.TrimSpace(raw)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func parseIntOrDefault(raw string, def int) int {
	s := strings.TrimSpace(raw)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func parseFloat01(raw string, def float64) float64 {
	s := strings.TrimSpace(raw)
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func envOrDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func webSearchEnabled(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "off", "none", "disabled":
		return false
	default:
		return true
	}
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}


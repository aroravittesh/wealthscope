package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("WEALTHSCOPE_SESSION_TTL", "")
	t.Setenv("INTENT_MIN_CONFIDENCE", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "9000" {
		t.Fatalf("port default mismatch: %s", cfg.Server.Port)
	}
	if cfg.Session.TTL != 24*time.Hour {
		t.Fatalf("ttl default mismatch: %s", cfg.Session.TTL)
	}
	if cfg.Intent.MinConfidence != 0 {
		t.Fatalf("min confidence default mismatch: %f", cfg.Intent.MinConfidence)
	}
}

func TestLoadFromEnv_ParsesAndClamps(t *testing.T) {
	t.Setenv("PORT", "9100")
	t.Setenv("INTENT_MIN_CONFIDENCE", "1.7")
	t.Setenv("WEALTHSCOPE_SESSION_TTL", "2h")
	t.Setenv("WEALTHSCOPE_WEB_SEARCH_PROVIDER", "off")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "9100" {
		t.Fatalf("port mismatch: %s", cfg.Server.Port)
	}
	if cfg.Intent.MinConfidence != 1 {
		t.Fatalf("expected clamp=1 got %f", cfg.Intent.MinConfidence)
	}
	if cfg.Session.TTL != 2*time.Hour {
		t.Fatalf("ttl mismatch: %s", cfg.Session.TTL)
	}
	if cfg.WebSearch.Enabled {
		t.Fatal("websearch should be disabled")
	}
}

func TestSafeSummary_NoSecrets(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "secret")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	s := cfg.SafeSummary()
	if strings.Contains(s, "secret") {
		t.Fatal("safe summary leaked secret")
	}
}


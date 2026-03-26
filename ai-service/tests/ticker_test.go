package tests

import (
    "testing"
    "wealthscope-ai/internal/ml"
)

func TestExtractTicker_DollarFormat(t *testing.T) {
    ticker := ml.ExtractTicker("What is the price of $AAPL today?")
    if ticker != "AAPL" {
        t.Fatalf("expected AAPL got %s", ticker)
    }
}

func TestExtractTicker_PlainFormat(t *testing.T) {
    ticker := ml.ExtractTicker("Tell me about TSLA stock")
    if ticker != "TSLA" {
        t.Fatalf("expected TSLA got %s", ticker)
    }
}

func TestExtractTicker_NoTicker(t *testing.T) {
    ticker := ml.ExtractTicker("What is a stock market?")
    if ticker != "" {
        t.Fatalf("expected empty got %s", ticker)
    }
}

func TestExtractTicker_IgnoresCommonWords(t *testing.T) {
    ticker := ml.ExtractTicker("What is the market doing?")
    if ticker == "THE" || ticker == "IS" {
        t.Fatalf("should not return common word as ticker got %s", ticker)
    }
}

func TestExtractTicker_DollarTakesPriority(t *testing.T) {
    ticker := ml.ExtractTicker("Compare $AAPL and MSFT")
    if ticker != "AAPL" {
        t.Fatalf("expected AAPL got %s", ticker)
    }
}
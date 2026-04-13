package tests

import (
	"testing"

	"wealthscope-ai/internal/entity"
)

func TestExtract_DollarTicker(t *testing.T) {
	r := entity.Extract("What is the price of $AAPL today?")
	if r.PrimaryTicker != "AAPL" {
		t.Fatalf("PrimaryTicker: want AAPL, got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 0 {
		t.Fatalf("SecondaryTickers: want empty, got %v", r.SecondaryTickers)
	}
}

func TestExtract_PlainTSLA(t *testing.T) {
	r := entity.Extract("Tell me about TSLA stock")
	if r.PrimaryTicker != "TSLA" {
		t.Fatalf("PrimaryTicker: want TSLA, got %q", r.PrimaryTicker)
	}
}

func TestExtract_CompanyNameApple(t *testing.T) {
	r := entity.Extract("How is Apple doing lately?")
	if r.PrimaryTicker != "AAPL" {
		t.Fatalf("PrimaryTicker: want AAPL, got %q", r.PrimaryTicker)
	}
	if len(r.CompanyMatches) != 1 || r.CompanyMatches[0] != "Apple" {
		t.Fatalf("CompanyMatches: want [Apple], got %v", r.CompanyMatches)
	}
}

func TestExtract_MultipleAppleAndMicrosoft(t *testing.T) {
	r := entity.Extract("Compare Apple and Microsoft revenue")
	if r.PrimaryTicker != "AAPL" {
		t.Fatalf("PrimaryTicker: want AAPL, got %q", r.PrimaryTicker)
	}
	wantSec := []string{"MSFT"}
	if len(r.SecondaryTickers) != len(wantSec) || r.SecondaryTickers[0] != wantSec[0] {
		t.Fatalf("SecondaryTickers: want %v, got %v", wantSec, r.SecondaryTickers)
	}
}

func TestExtract_NoTicker(t *testing.T) {
	r := entity.Extract("What is a stock market?")
	if r.PrimaryTicker != "" {
		t.Fatalf("PrimaryTicker: want empty, got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 0 {
		t.Fatalf("SecondaryTickers: want empty, got %v", r.SecondaryTickers)
	}
}

func TestExtract_FalsePositivePrevention(t *testing.T) {
	r := entity.Extract("What is the market doing?")
	if r.PrimaryTicker == "THE" || r.PrimaryTicker == "IS" {
		t.Fatalf("should not use common word as ticker, got %q", r.PrimaryTicker)
	}
	r2 := entity.Extract("WHAT IS THE MARKET DOING TODAY")
	if r2.PrimaryTicker != "" {
		t.Fatalf("all-caps prose should not yield a ticker, got %q", r2.PrimaryTicker)
	}
}

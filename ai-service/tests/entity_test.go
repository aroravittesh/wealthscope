package tests

import (
	"testing"

	"wealthscope-ai/internal/entity"
)

// --- Existing baseline tests (kept) -----------------------------------------

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
	if len(r.SecondaryTickers) != 1 || r.SecondaryTickers[0] != "MSFT" {
		t.Fatalf("SecondaryTickers: want [MSFT], got %v", r.SecondaryTickers)
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

// --- New: alias / multi-token / multi-entity --------------------------------

func TestExtract_AliasAlphabet(t *testing.T) {
	r := entity.Extract("Is Alphabet a good long-term hold?")
	if r.PrimaryTicker != "GOOGL" {
		t.Fatalf("Alphabet must resolve to GOOGL, got %q", r.PrimaryTicker)
	}
	if len(r.CompanyMatches) == 0 || r.CompanyMatches[0] != "Alphabet" {
		t.Fatalf("CompanyMatches: want [Alphabet], got %v", r.CompanyMatches)
	}
}

func TestExtract_MultiTokenBankOfAmerica(t *testing.T) {
	r := entity.Extract("How risky is Bank of America right now?")
	if r.PrimaryTicker != "BAC" {
		t.Fatalf("Bank of America must resolve to BAC, got %q", r.PrimaryTicker)
	}
	if len(r.CompanyMatches) == 0 || r.CompanyMatches[0] != "Bank of America" {
		t.Fatalf("CompanyMatches: want [Bank of America], got %v", r.CompanyMatches)
	}
}

func TestExtract_LongestPhraseWins(t *testing.T) {
	// "Berkshire Hathaway" must consume both tokens before standalone "berkshire" can match.
	r := entity.Extract("What is Berkshire Hathaway up to?")
	if r.PrimaryTicker != "BRK.B" {
		t.Fatalf("primary: want BRK.B, got %q", r.PrimaryTicker)
	}
	// We should see exactly one company label, not double-counted.
	count := 0
	for _, l := range r.CompanyMatches {
		if l == "Berkshire Hathaway" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Berkshire Hathaway should appear once, got %d in %v", count, r.CompanyMatches)
	}
}

func TestExtract_HyphenatedCocaCola(t *testing.T) {
	r := entity.Extract("Compare Coca-Cola with PepsiCo")
	if r.PrimaryTicker != "KO" {
		t.Fatalf("Coca-Cola must resolve to KO, got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 1 || r.SecondaryTickers[0] != "PEP" {
		t.Fatalf("SecondaryTickers: want [PEP], got %v", r.SecondaryTickers)
	}
}

func TestExtract_MultiEntityComparison(t *testing.T) {
	r := entity.Extract("Is Tesla more volatile than Nvidia?")
	if r.PrimaryTicker != "TSLA" {
		t.Fatalf("primary should be TSLA (sentence-first), got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 1 || r.SecondaryTickers[0] != "NVDA" {
		t.Fatalf("secondary: want [NVDA], got %v", r.SecondaryTickers)
	}
}

func TestExtract_RankingPrefersSentenceOrder(t *testing.T) {
	// Mixed signals: Tesla (dict) appears before AAPL (plain). Sentence order wins.
	r := entity.Extract("Compare Tesla and AAPL")
	if r.PrimaryTicker != "TSLA" {
		t.Fatalf("sentence order should make TSLA primary, got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 1 || r.SecondaryTickers[0] != "AAPL" {
		t.Fatalf("secondary: want [AAPL], got %v", r.SecondaryTickers)
	}
}

// --- New: confidence & agreement -------------------------------------------

func TestExtract_AgreementBoostsConfidence(t *testing.T) {
	plain := entity.Extract("AAPL outlook")
	combined := entity.Extract("Apple ($AAPL) outlook")
	if combined.PrimaryTicker != "AAPL" {
		t.Fatalf("primary should be AAPL, got %q", combined.PrimaryTicker)
	}
	if combined.Confidence <= plain.Confidence {
		t.Fatalf("multi-source agreement must boost confidence: plain=%f combined=%f",
			plain.Confidence, combined.Confidence)
	}
	if combined.Confidence > 1.0 {
		t.Fatalf("confidence must remain in [0,1], got %f", combined.Confidence)
	}
}

func TestExtract_DictOnlyHasLowerConfidenceThanDollar(t *testing.T) {
	dict := entity.Extract("How is Apple doing?")
	dollar := entity.Extract("How is $AAPL doing?")
	if !(dollar.Confidence > dict.Confidence) {
		t.Fatalf("$AAPL should outscore dict-only Apple: dollar=%f dict=%f",
			dollar.Confidence, dict.Confidence)
	}
}

// --- New: false-positive prevention ----------------------------------------

func TestExtract_NoFalsePositiveOnFinanceJargon(t *testing.T) {
	cases := []string{
		"Looking at the IPO calendar this week",
		"What is a good ETF for tech exposure?",
		"How does CEO compensation affect EPS?",
		"GDP growth in the USA was strong",
		"Is the FED about to cut rates?",
	}
	for _, msg := range cases {
		r := entity.Extract(msg)
		if r.PrimaryTicker != "" {
			t.Fatalf("expected no ticker for %q, got %q", msg, r.PrimaryTicker)
		}
	}
}

func TestExtract_PossessiveStillResolves(t *testing.T) {
	r := entity.Extract("Apple's earnings beat estimates")
	if r.PrimaryTicker != "AAPL" {
		t.Fatalf("possessive Apple's should resolve to AAPL, got %q", r.PrimaryTicker)
	}
}

// --- New: well-known query patterns ----------------------------------------

func TestExtract_LatestNewsOnGoogle(t *testing.T) {
	r := entity.Extract("What is the latest news on Google?")
	if r.PrimaryTicker != "GOOGL" {
		t.Fatalf("Google must resolve to GOOGL, got %q", r.PrimaryTicker)
	}
}

func TestExtract_RiskOfMetaVsAmazon(t *testing.T) {
	r := entity.Extract("How risky is Meta compared to Amazon?")
	if r.PrimaryTicker != "META" {
		t.Fatalf("primary should be META, got %q", r.PrimaryTicker)
	}
	if len(r.SecondaryTickers) != 1 || r.SecondaryTickers[0] != "AMZN" {
		t.Fatalf("secondary: want [AMZN], got %v", r.SecondaryTickers)
	}
}

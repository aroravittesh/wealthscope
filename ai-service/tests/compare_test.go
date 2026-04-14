package tests

import (
	"errors"
	"strings"
	"testing"

	"wealthscope-ai/internal/compare"
	"wealthscope-ai/internal/market"
)

type mockFetcher struct {
	quotes      map[string]*market.GlobalQuote
	overviews   map[string]*market.CompanyOverview
	quoteErr    map[string]error
	overviewErr map[string]error
}

func (m mockFetcher) Quote(symbol string) (*market.GlobalQuote, error) {
	if e := m.quoteErr[symbol]; e != nil {
		return nil, e
	}
	return m.quotes[symbol], nil
}

func (m mockFetcher) Overview(symbol string) (*market.CompanyOverview, error) {
	if e := m.overviewErr[symbol]; e != nil {
		return nil, e
	}
	return m.overviews[symbol], nil
}

func TestNormalizeAndValidate_TwoSymbols(t *testing.T) {
	got, err := compare.NormalizeAndValidate([]string{"aapl", " msft "})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "AAPL" || got[1] != "MSFT" {
		t.Fatalf("got %v", got)
	}
}

func TestNormalizeAndValidate_FourSymbols(t *testing.T) {
	_, err := compare.NormalizeAndValidate([]string{"A", "B", "C", "D"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeAndValidate_TooFew(t *testing.T) {
	_, err := compare.NormalizeAndValidate([]string{"AAPL"})
	if !errors.Is(err, compare.ErrInvalidSymbolCount) {
		t.Fatalf("want ErrInvalidSymbolCount got %v", err)
	}
}

func TestNormalizeAndValidate_TooMany(t *testing.T) {
	_, err := compare.NormalizeAndValidate([]string{"A", "B", "C", "D", "E"})
	if !errors.Is(err, compare.ErrInvalidSymbolCount) {
		t.Fatalf("want ErrInvalidSymbolCount got %v", err)
	}
}

func TestNormalizeAndValidate_EmptyString(t *testing.T) {
	_, err := compare.NormalizeAndValidate([]string{"AAPL", "  "})
	if !errors.Is(err, compare.ErrEmptySymbol) {
		t.Fatalf("want ErrEmptySymbol got %v", err)
	}
}

func TestNormalizeAndValidate_DuplicatesCollapseToOne(t *testing.T) {
	_, err := compare.NormalizeAndValidate([]string{"AAPL", "AAPL"})
	if !errors.Is(err, compare.ErrInvalidSymbolCount) {
		t.Fatalf("want ErrInvalidSymbolCount got %v", err)
	}
}

func TestNormalizeAndValidate_MissingSymbolsNil(t *testing.T) {
	_, err := compare.NormalizeAndValidate(nil)
	if !errors.Is(err, compare.ErrInvalidSymbolCount) {
		t.Fatalf("want ErrInvalidSymbolCount got %v", err)
	}
}

func TestCompare_MockSuccess(t *testing.T) {
	f := mockFetcher{
		quotes: map[string]*market.GlobalQuote{
			"AAPL": {Symbol: "AAPL", Price: "180.00"},
			"MSFT": {Symbol: "MSFT", Price: "400.00"},
		},
		overviews: map[string]*market.CompanyOverview{
			"AAPL": {Name: "Apple", Sector: "Technology", Industry: "Consumer Electronics", MarketCap: "3000000000000", PERatio: "28", Beta: "1.2"},
			"MSFT": {Name: "Microsoft", Sector: "Technology", Industry: "Software", MarketCap: "2800000000000", PERatio: "35", Beta: "0.9"},
		},
	}
	resp, err := compare.Compare(f, []string{"AAPL", "MSFT"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Comparisons) != 2 {
		t.Fatalf("comparisons len %d", len(resp.Comparisons))
	}
	if resp.Comparisons[0].Price != "180.00" || resp.Comparisons[0].Beta != "1.2" {
		t.Fatalf("first row %+v", resp.Comparisons[0])
	}
	if resp.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(resp.Summary) < 20 {
		t.Fatalf("summary too short: %q", resp.Summary)
	}
}

func TestCompare_QuoteFailure(t *testing.T) {
	f := mockFetcher{
		quotes: map[string]*market.GlobalQuote{
			"MSFT": {Price: "1"},
		},
		quoteErr: map[string]error{
			"AAPL": errors.New("not found"),
		},
		overviews: map[string]*market.CompanyOverview{
			"AAPL": {Beta: "1"},
			"MSFT": {Beta: "1"},
		},
	}
	_, err := compare.Compare(f, []string{"AAPL", "MSFT"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompare_SummaryNoBuySellLanguage(t *testing.T) {
	f := mockFetcher{
		quotes: map[string]*market.GlobalQuote{
			"AAPL": {Price: "1"},
			"MSFT": {Price: "2"},
		},
		overviews: map[string]*market.CompanyOverview{
			"AAPL": {Sector: "Technology", Beta: "1.1", PERatio: "22", MarketCap: "1000"},
			"MSFT": {Sector: "Technology", Beta: "1.0", PERatio: "30", MarketCap: "2000"},
		},
	}
	resp, err := compare.Compare(f, []string{"AAPL", "MSFT"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Summary, "not buy or sell") {
		t.Fatalf("summary should disclaim recommendations: %q", resp.Summary)
	}
}

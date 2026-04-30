package entity

import (
	"regexp"
	"sort"
	"strings"
)

// Alias is one company name (possibly multi-token) that resolves to a ticker.
// Phrase tokens are lowercase; matching is whole-word against a normalized
// (lowercased, hyphens → spaces) copy of the input message.
type Alias struct {
	Phrase []string
	Ticker string
	Label  string
}

// companyAliases is the curated alias dictionary. Single-word risky names
// (Target, Visa, Ford, etc.) are intentionally omitted to limit false positives.
var companyAliases = []Alias{
	// Mega-cap tech
	{Phrase: []string{"alphabet"}, Ticker: "GOOGL", Label: "Alphabet"},
	{Phrase: []string{"google"}, Ticker: "GOOGL", Label: "Google"},
	{Phrase: []string{"apple"}, Ticker: "AAPL", Label: "Apple"},
	{Phrase: []string{"microsoft"}, Ticker: "MSFT", Label: "Microsoft"},
	{Phrase: []string{"meta"}, Ticker: "META", Label: "Meta"},
	{Phrase: []string{"facebook"}, Ticker: "META", Label: "Facebook"},
	{Phrase: []string{"amazon"}, Ticker: "AMZN", Label: "Amazon"},
	{Phrase: []string{"netflix"}, Ticker: "NFLX", Label: "Netflix"},
	{Phrase: []string{"tesla"}, Ticker: "TSLA", Label: "Tesla"},
	{Phrase: []string{"nvidia"}, Ticker: "NVDA", Label: "NVIDIA"},
	{Phrase: []string{"intel"}, Ticker: "INTC", Label: "Intel"},
	{Phrase: []string{"amd"}, Ticker: "AMD", Label: "AMD"},
	{Phrase: []string{"oracle"}, Ticker: "ORCL", Label: "Oracle"},
	{Phrase: []string{"salesforce"}, Ticker: "CRM", Label: "Salesforce"},
	{Phrase: []string{"adobe"}, Ticker: "ADBE", Label: "Adobe"},
	{Phrase: []string{"ibm"}, Ticker: "IBM", Label: "IBM"},
	{Phrase: []string{"cisco"}, Ticker: "CSCO", Label: "Cisco"},
	{Phrase: []string{"qualcomm"}, Ticker: "QCOM", Label: "Qualcomm"},
	{Phrase: []string{"broadcom"}, Ticker: "AVGO", Label: "Broadcom"},
	{Phrase: []string{"paypal"}, Ticker: "PYPL", Label: "PayPal"},

	// Financials
	{Phrase: []string{"berkshire", "hathaway"}, Ticker: "BRK.B", Label: "Berkshire Hathaway"},
	{Phrase: []string{"berkshire"}, Ticker: "BRK.B", Label: "Berkshire Hathaway"},
	{Phrase: []string{"jp", "morgan"}, Ticker: "JPM", Label: "JPMorgan"},
	{Phrase: []string{"jpmorgan"}, Ticker: "JPM", Label: "JPMorgan"},
	{Phrase: []string{"goldman", "sachs"}, Ticker: "GS", Label: "Goldman Sachs"},
	{Phrase: []string{"goldman"}, Ticker: "GS", Label: "Goldman Sachs"},
	{Phrase: []string{"morgan", "stanley"}, Ticker: "MS", Label: "Morgan Stanley"},
	{Phrase: []string{"bank", "of", "america"}, Ticker: "BAC", Label: "Bank of America"},
	{Phrase: []string{"wells", "fargo"}, Ticker: "WFC", Label: "Wells Fargo"},
	{Phrase: []string{"citigroup"}, Ticker: "C", Label: "Citigroup"},
	{Phrase: []string{"mastercard"}, Ticker: "MA", Label: "Mastercard"},

	// Healthcare / pharma
	{Phrase: []string{"johnson", "and", "johnson"}, Ticker: "JNJ", Label: "Johnson & Johnson"},
	{Phrase: []string{"johnson", "johnson"}, Ticker: "JNJ", Label: "Johnson & Johnson"},
	{Phrase: []string{"eli", "lilly"}, Ticker: "LLY", Label: "Eli Lilly"},
	{Phrase: []string{"pfizer"}, Ticker: "PFE", Label: "Pfizer"},
	{Phrase: []string{"merck"}, Ticker: "MRK", Label: "Merck"},
	{Phrase: []string{"unitedhealth"}, Ticker: "UNH", Label: "UnitedHealth"},

	// Consumer / retail
	{Phrase: []string{"walt", "disney"}, Ticker: "DIS", Label: "Disney"},
	{Phrase: []string{"disney"}, Ticker: "DIS", Label: "Disney"},
	{Phrase: []string{"coca", "cola"}, Ticker: "KO", Label: "Coca-Cola"},
	{Phrase: []string{"pepsico"}, Ticker: "PEP", Label: "PepsiCo"},
	{Phrase: []string{"pepsi"}, Ticker: "PEP", Label: "PepsiCo"},
	{Phrase: []string{"walmart"}, Ticker: "WMT", Label: "Walmart"},
	{Phrase: []string{"costco"}, Ticker: "COST", Label: "Costco"},
	{Phrase: []string{"home", "depot"}, Ticker: "HD", Label: "Home Depot"},
	{Phrase: []string{"mcdonald", "s"}, Ticker: "MCD", Label: "McDonald's"},
	{Phrase: []string{"mcdonalds"}, Ticker: "MCD", Label: "McDonald's"},
	{Phrase: []string{"starbucks"}, Ticker: "SBUX", Label: "Starbucks"},
	{Phrase: []string{"boeing"}, Ticker: "BA", Label: "Boeing"},
	{Phrase: []string{"general", "motors"}, Ticker: "GM", Label: "General Motors"},

	// Crypto-adjacent / fintech that often comes up in chat
	{Phrase: []string{"coinbase"}, Ticker: "COIN", Label: "Coinbase"},
	{Phrase: []string{"shopify"}, Ticker: "SHOP", Label: "Shopify"},
	{Phrase: []string{"uber"}, Ticker: "UBER", Label: "Uber"},
	{Phrase: []string{"airbnb"}, Ticker: "ABNB", Label: "Airbnb"},
}

// aliasMatcher precompiles each alias as a whole-word regex against normalized text.
type aliasMatcher struct {
	alias Alias
	re    *regexp.Regexp
}

var compiledAliases []aliasMatcher

func init() {
	// Sort longest-phrase first so multi-token aliases consume their tokens
	// before shorter aliases can match a sub-word (e.g. "berkshire hathaway"
	// before "berkshire").
	sort.SliceStable(companyAliases, func(i, j int) bool {
		return len(companyAliases[i].Phrase) > len(companyAliases[j].Phrase)
	})
	compiledAliases = make([]aliasMatcher, 0, len(companyAliases))
	for _, a := range companyAliases {
		compiledAliases = append(compiledAliases, aliasMatcher{
			alias: a,
			re:    buildAliasRegex(a.Phrase),
		})
	}
}

// buildAliasRegex builds a `\bp1\s+p2\s+...\b`-style regex.
// Each phrase token is regexp-escaped so symbols like "BRK.B" never appear here
// (those are tickers, not phrase tokens), but escaping is a safety belt.
func buildAliasRegex(phrase []string) *regexp.Regexp {
	parts := make([]string, len(phrase))
	for i, p := range phrase {
		parts[i] = regexp.QuoteMeta(p)
	}
	pattern := `(?i)\b` + strings.Join(parts, `\s+`) + `\b`
	return regexp.MustCompile(pattern)
}

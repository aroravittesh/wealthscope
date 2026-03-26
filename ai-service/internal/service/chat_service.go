package service

import (
    "fmt"
    "strings"

    "wealthscope-ai/internal/market"
    "wealthscope-ai/internal/ml"
    "wealthscope-ai/internal/openai"
)

func ProcessMessage(sessionID string, message string) (string, error) {
    // Step 1: detect intent and extract ticker
    intentResult := ml.DetectIntent(message)

    // Step 2: analyze sentiment
    sentiment := ml.AnalyzeSentiment(message)

    // Step 3: build enriched prompt
    enriched := message

    // Step 4: inject live market data if ticker found
    if intentResult.Ticker != "" {
        ticker := intentResult.Ticker

        quote, err := market.GetStockQuote(ticker)
        if err == nil {
            enriched += fmt.Sprintf(
                "\n\n[Live Price Data for %s]\nPrice: $%s | High: $%s | Low: $%s | Change: %s (%s) | Volume: %s",
                quote.Symbol, quote.Price, quote.High, quote.Low,
                quote.Change, quote.ChangePercent, quote.Volume,
            )
        }

        overview, err := market.GetCompanyOverview(ticker)
        if err == nil {
            enriched += fmt.Sprintf(
                "\n\n[Company Fundamentals]\nName: %s | Sector: %s | Industry: %s\nMarket Cap: $%s | P/E: %s | EPS: %s | Beta: %s\n52W High: $%s | 52W Low: $%s | Dividend Yield: %s | Profit Margin: %s\nDescription: %s",
                overview.Name, overview.Sector, overview.Industry,
                overview.MarketCap, overview.PERatio, overview.EPS, overview.Beta,
                overview.Week52High, overview.Week52Low,
                overview.DivYield, overview.ProfitMargin,
                overview.Description,
            )
        }

        news, err := market.GetMarketNews(ticker)
        if err == nil && len(news) > 0 {
            enriched += fmt.Sprintf("\n\n[Latest News for %s]", ticker)
            for i, article := range news {
                if i >= 3 {
                    break
                }
                enriched += fmt.Sprintf(
                    "\n%d. %s — %s (%s)",
                    i+1, article.Title, article.Source.Name, article.PublishedAt,
                )
            }
        }
    }

    // Step 5: inject system context
    enriched += fmt.Sprintf(
        "\n\n[System Context]\nIntent: %s | Ticker: %s | Sentiment: %s | Confidence: %.2f",
        intentResult.Intent,
        strings.ToUpper(intentResult.Ticker),
        sentiment,
        intentResult.Confidence,
    )

    return openai.CallOpenAI(sessionID, enriched)
}

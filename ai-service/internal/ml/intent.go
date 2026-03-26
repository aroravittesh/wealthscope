package ml

import "strings"

type Intent string

const (
    IntentStockPrice     Intent = "STOCK_PRICE"
    IntentRiskAnalysis   Intent = "RISK_ANALYSIS"
    IntentMarketNews     Intent = "MARKET_NEWS"
    IntentPortfolioTip   Intent = "PORTFOLIO_TIP"
    IntentGeneralMarket  Intent = "GENERAL_MARKET"
    IntentUnknown        Intent = "UNKNOWN"
)

type IntentResult struct {
    Intent     Intent
    Ticker     string
    Confidence float64
}

// keyword maps for each intent
var intentKeywords = map[Intent][]string{
    IntentStockPrice:    {"price", "trading at", "current price", "stock price", "how much is", "quote"},
    IntentRiskAnalysis:  {"risk", "volatile", "volatility", "safe", "dangerous", "should i buy", "analysis"},
    IntentMarketNews:    {"news", "latest", "today", "happening", "update", "recent"},
    IntentPortfolioTip:  {"portfolio", "diversify", "allocate", "holdings", "invest", "suggestion", "recommend"},
    IntentGeneralMarket: {"market", "s&p", "nasdaq", "dow", "index", "bull", "bear", "recession"},
}

func DetectIntent(message string) IntentResult {
    lower := strings.ToLower(message)
    best := IntentResult{Intent: IntentUnknown, Confidence: 0.0}

    for intent, keywords := range intentKeywords {
        score := 0.0
        for _, kw := range keywords {
            if strings.Contains(lower, kw) {
                score += 1.0 / float64(len(keywords))
            }
        }
        if score > best.Confidence {
            best.Intent = intent
            best.Confidence = score
        }
    }

    ticker := ExtractTicker(message)
    best.Ticker = ticker
    return best
}
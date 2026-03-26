package tests

import (
    "testing"
    "wealthscope-ai/internal/ml"
)

func TestDetectIntent_StockPrice(t *testing.T) {
    result := ml.DetectIntent("What is the current price of AAPL?")
    if result.Intent != ml.IntentStockPrice {
        t.Fatalf("expected STOCK_PRICE got %s", result.Intent)
    }
}

func TestDetectIntent_RiskAnalysis(t *testing.T) {
    result := ml.DetectIntent("Is TSLA a volatile stock?")
    if result.Intent != ml.IntentRiskAnalysis {
        t.Fatalf("expected RISK_ANALYSIS got %s", result.Intent)
    }
}

func TestDetectIntent_MarketNews(t *testing.T) {
    result := ml.DetectIntent("What is the latest news on MSFT?")
    if result.Intent != ml.IntentMarketNews {
        t.Fatalf("expected MARKET_NEWS got %s", result.Intent)
    }
}

func TestDetectIntent_PortfolioTip(t *testing.T) {
    result := ml.DetectIntent("How should I diversify my portfolio?")
    if result.Intent != ml.IntentPortfolioTip {
        t.Fatalf("expected PORTFOLIO_TIP got %s", result.Intent)
    }
}

func TestDetectIntent_GeneralMarket(t *testing.T) {
    result := ml.DetectIntent("How is the S&P 500 index doing in this bull market?")
    if result.Intent != ml.IntentGeneralMarket {
        t.Fatalf("expected GENERAL_MARKET got %s", result.Intent)
    }
}

func TestDetectIntent_Unknown(t *testing.T) {
    result := ml.DetectIntent("Hello how are you?")
    if result.Intent != ml.IntentUnknown {
        t.Fatalf("expected UNKNOWN got %s", result.Intent)
    }
}

func TestDetectIntent_TickerExtracted(t *testing.T) {
    result := ml.DetectIntent("What is the price of $AAPL?")
    if result.Ticker != "AAPL" {
        t.Fatalf("expected ticker AAPL got %s", result.Ticker)
    }
}

func TestDetectIntent_ConfidenceNonZero(t *testing.T) {
    result := ml.DetectIntent("What is the stock price?")
    if result.Confidence <= 0 {
        t.Fatalf("expected confidence > 0 got %f", result.Confidence)
    }
}
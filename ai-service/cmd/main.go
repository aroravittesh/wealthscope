package main

import (
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "wealthscope-ai/internal/handler"
    "wealthscope-ai/internal/market"
    "wealthscope-ai/internal/ml"
)

func main() {

    godotenv.Load()

    router := gin.Default()

    // CORS middleware
    router.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "AI Service running",
        })
    })

    // Chat endpoint
    // uses intent detection + sentiment + market data injection
    // internally calls ml.DetectIntent, ml.AnalyzeSentiment, market.GetStockQuote
    router.POST("/chat", handler.ChatHandler)

    // Risk scoring endpoint
    // accepts a list of holdings with symbol, allocation, beta
    // returns a risk report with score, level, and explanation
    router.POST("/risk", func(c *gin.Context) {
        var body struct {
            Holdings []ml.PortfolioHolding `json:"holdings"`
        }
        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }
        report := ml.ScorePortfolio(body.Holdings)
        c.JSON(http.StatusOK, report)
    })

    // Stock quote endpoint
    // returns real-time price, change, volume for a given symbol
    router.GET("/quote/:symbol", func(c *gin.Context) {
        symbol := c.Param("symbol")
        quote, err := market.GetStockQuote(symbol)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, quote)
    })

    // Company overview endpoint
    // returns name, sector, industry, market cap, P/E, beta, 52W high/low
    router.GET("/company/:symbol", func(c *gin.Context) {
        symbol := c.Param("symbol")
        overview, err := market.GetCompanyOverview(symbol)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, overview)
    })

    // Sentiment analysis endpoint
    // accepts a text string and returns BULLISH, BEARISH, or NEUTRAL
    router.POST("/sentiment", func(c *gin.Context) {
        var body struct {
            Text string `json:"text"`
        }
        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }
        if body.Text == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Text cannot be empty"})
            return
        }
        sentiment := ml.AnalyzeSentiment(body.Text)
        c.JSON(http.StatusOK, gin.H{"sentiment": sentiment})
    })

    // Intent detection endpoint
    // accepts a message and returns detected intent, ticker, and confidence
    router.POST("/intent", func(c *gin.Context) {
        var body struct {
            Message string `json:"message"`
        }
        if err := c.BindJSON(&body); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }
        if body.Message == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
            return
        }
        result := ml.DetectIntent(body.Message)
        c.JSON(http.StatusOK, result)
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "9000"
    }

    router.Run(":" + port)
}
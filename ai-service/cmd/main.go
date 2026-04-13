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
    router.SetTrustedProxies(nil)

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
    router.POST("/chat", handler.ChatHandler)

    router.POST("/predict/risk-drift", handler.RiskDriftHandler)

    router.POST("/portfolio/explain", handler.PortfolioExplainHandler)

    router.POST("/compare", handler.CompareHandler)

    router.GET("/news-sentiment/:symbol", handler.NewsSentimentHandler)

    // Risk scoring endpoint
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
    router.GET("/company/:symbol", func(c *gin.Context) {
        symbol := c.Param("symbol")
        overview, err := market.GetCompanyOverview(symbol)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, overview)
    })
// News endpoint
router.GET("/news/:symbol", func(c *gin.Context) {
    symbol := c.Param("symbol")
    news, err := market.GetMarketNews(symbol)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if len(news) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "No news found for " + symbol})
        return
    }
    c.JSON(http.StatusOK, gin.H{"symbol": symbol, "news": news})
})

// Clear session endpoint
router.DELETE("/chat/session/:session_id", handler.ClearChatHandler)
    // Sentiment analysis endpoint
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
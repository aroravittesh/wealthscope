package main

import (
	"log"
	"net/http"

	"wealthscope-ai/internal/config"
	"wealthscope-ai/internal/feedback"
	"wealthscope-ai/internal/handler"
	"wealthscope-ai/internal/market"
	"wealthscope-ai/internal/ml"
	"wealthscope-ai/internal/openai"
	"wealthscope-ai/internal/rag"
	"wealthscope-ai/internal/websearch"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

    _ = godotenv.Load()
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	// Startup wiring from centralized config.
	openai.SetAPIKey(cfg.OpenAI.APIKey)
	openai.SetDefaultStore(openai.NewStore(openai.StoreConfig{
		TTL:              cfg.Session.TTL,
		MaxMessages:      cfg.Session.MaxMessages,
		KeepAfterCompact: cfg.Session.KeepAfterCompact,
	}))
	market.SetConfig(market.Config{
		AlphaVantageAPIKey: cfg.Market.AlphaVantageAPIKey,
		NewsAPIKey:         cfg.Market.NewsAPIKey,
	})
	ml.SetDefaultIntentConfig(ml.IntentConfig{
		ClassifierBaseURL: cfg.Intent.ClassifierURL,
		Client:            &http.Client{Timeout: cfg.Intent.Timeout},
		MinConfidence:     cfg.Intent.MinConfidence,
	})
	rag.SetQADatasetPath(cfg.RAG.QADatasetPath)
	feedback.SetDefaultStore(feedback.NewJSONLStore(cfg.Feedback.Path))
	websearch.SetDefaultProviderFromConfig(websearch.ProviderConfig{
		Provider:  cfg.WebSearch.Provider,
		TavilyKey: cfg.WebSearch.TavilyAPIKey,
		Timeout:   cfg.WebSearch.Timeout,
	})
	log.Printf("config loaded: %s", cfg.SafeSummary())

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
            handler.RespondSuccess(c, 200, "AI Service running", gin.H{
                "status": "AI Service running",
            })
    })

    // Chat endpoint
    router.POST("/chat", handler.ChatHandler)

    router.POST("/predict/risk-drift", handler.RiskDriftHandler)

    router.POST("/portfolio/explain", handler.PortfolioExplainHandler)
    router.POST("/portfolio/summarize", handler.PortfolioSummarizeHandler)
    router.POST("/portfolio/changes", handler.PortfolioChangesHandler)

    router.POST("/compare", handler.CompareHandler)

    router.GET("/news-sentiment/:symbol", handler.NewsSentimentHandler)

    // Feedback collection (learning-oriented logging pipeline).
    router.POST("/feedback", handler.RecordFeedbackHandler)
    router.GET("/feedback", handler.ListFeedbackHandler)
    router.GET("/feedback/export", handler.ExportFeedbackHandler)

    // Risk scoring endpoint
    router.POST("/risk", func(c *gin.Context) {
        var body struct {
            Holdings []ml.PortfolioHolding `json:"holdings"`
        }
        if err := c.BindJSON(&body); err != nil {
            handler.RespondBadRequest(c, "Request failed", "Invalid request")
            return
        }
        report := ml.ScorePortfolio(body.Holdings)
        handler.RespondSuccess(c, 200, "Risk score generated", report)
    })

    // Stock quote endpoint
    router.GET("/quote/:symbol", func(c *gin.Context) {
        symbol := c.Param("symbol")
        quote, err := market.GetStockQuote(symbol)
        if err != nil {
            handler.RespondError(c, 404, "Request failed", err.Error())
            return
        }
        handler.RespondSuccess(c, 200, "Stock quote retrieved", quote)
    })

    // Company overview endpoint
    router.GET("/company/:symbol", func(c *gin.Context) {
        symbol := c.Param("symbol")
        overview, err := market.GetCompanyOverview(symbol)
        if err != nil {
            handler.RespondError(c, 404, "Request failed", err.Error())
            return
        }
        handler.RespondSuccess(c, 200, "Company overview retrieved", overview)
    })
// News endpoint
router.GET("/news/:symbol", func(c *gin.Context) {
    symbol := c.Param("symbol")
    news, err := market.GetMarketNews(symbol)
    if err != nil {
        handler.RespondError(c, 500, "Request failed", err.Error())
        return
    }
    if len(news) == 0 {
        handler.RespondError(c, 404, "Request failed", "No news found for "+symbol)
        return
    }
    handler.RespondSuccess(c, 200, "Market news retrieved", gin.H{"symbol": symbol, "news": news})
})

// Clear session endpoint
router.DELETE("/chat/session/:session_id", handler.ClearChatHandler)
    // Sentiment analysis endpoint
    router.POST("/sentiment", func(c *gin.Context) {
        var body struct {
            Text string `json:"text"`
        }
        if err := c.BindJSON(&body); err != nil {
            handler.RespondBadRequest(c, "Request failed", "Invalid request")
            return
        }
        if body.Text == "" {
            handler.RespondBadRequest(c, "Request failed", "Text cannot be empty")
            return
        }
        sentiment := ml.AnalyzeSentiment(body.Text)
        handler.RespondSuccess(c, 200, "Sentiment generated", gin.H{"sentiment": sentiment})
    })

    // Intent detection endpoint
    router.POST("/intent", func(c *gin.Context) {
        var body struct {
            Message string `json:"message"`
        }
        if err := c.BindJSON(&body); err != nil {
            handler.RespondBadRequest(c, "Request failed", "Invalid request")
            return
        }
        if body.Message == "" {
            handler.RespondBadRequest(c, "Request failed", "Message cannot be empty")
            return
        }
        result := ml.DetectIntent(body.Message)
        handler.RespondSuccess(c, 200, "Intent detected", result)
    })

    router.Run(":" + cfg.Server.Port)
}

package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/newsentiment"
)

// NewsSentimentHandler handles GET /news-sentiment/:symbol.
func NewsSentimentHandler(c *gin.Context) {
	NewsSentimentHandlerWithFetcher(newsentiment.LiveFetcher{})(c)
}

// NewsSentimentHandlerWithFetcher allows injecting a news source (tests).
func NewsSentimentHandlerWithFetcher(fetcher newsentiment.NewsFetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.Param("symbol"))
		if raw == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
			return
		}
		symbol := strings.ToUpper(raw)

		articles, err := fetcher.FetchNews(symbol)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		resp := newsentiment.Aggregate(symbol, articles)
		c.JSON(http.StatusOK, resp)
	}
}

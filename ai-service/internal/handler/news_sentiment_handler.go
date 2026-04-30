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
		raw, ok := RequirePathParam(c, "symbol", "symbol required")
		if !ok {
			return
		}
		symbol := strings.ToUpper(raw)

		articles, err := fetcher.FetchNews(symbol)
		if err != nil {
			RespondError(c, http.StatusBadGateway, "Request failed", err.Error())
			return
		}

		resp := newsentiment.Aggregate(symbol, articles)
		RespondSuccess(c, http.StatusOK, "News sentiment generated", resp)
	}
}

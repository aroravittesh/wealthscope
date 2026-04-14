package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/compare"
)

// HTTPStatusForCompareError maps compare errors to HTTP status codes for POST /compare.
func HTTPStatusForCompareError(err error) int {
	switch {
	case errors.Is(err, compare.ErrInvalidSymbolCount), errors.Is(err, compare.ErrEmptySymbol):
		return http.StatusBadRequest
	default:
		msg := err.Error()
		if strings.Contains(msg, "quote:") || strings.Contains(msg, "overview:") {
			return http.StatusBadGateway
		}
		return http.StatusInternalServerError
	}
}

// CompareHandler handles POST /compare.
func CompareHandler(c *gin.Context) {
	CompareHandlerWithFetcher(compare.LiveFetcher{})(c)
}

// CompareHandlerWithFetcher allows injecting a fetcher (tests).
func CompareHandlerWithFetcher(fetcher compare.Fetcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req compare.Request
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		resp, err := compare.Compare(fetcher, req.Symbols)
		if err != nil {
			c.JSON(HTTPStatusForCompareError(err), gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

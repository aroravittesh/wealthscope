package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/compare"
)

// CompareHandler handles POST /compare.
func CompareHandler(c *gin.Context) {
	var req compare.Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	resp, err := compare.Compare(compare.LiveFetcher{}, req.Symbols)
	if err != nil {
		switch {
		case errors.Is(err, compare.ErrInvalidSymbolCount):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, compare.ErrEmptySymbol):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			if strings.Contains(err.Error(), "quote:") || strings.Contains(err.Error(), "overview:") {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

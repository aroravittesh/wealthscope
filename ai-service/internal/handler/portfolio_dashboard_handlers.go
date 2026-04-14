package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/portfoliosvc"
)

// PortfolioSummarizeHandler handles POST /portfolio/summarize.
func PortfolioSummarizeHandler(c *gin.Context) {
	var req portfoliosvc.SummarizeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	resp, err := portfoliosvc.Summarize(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PortfolioChangesHandler handles POST /portfolio/changes.
func PortfolioChangesHandler(c *gin.Context) {
	var req portfoliosvc.ChangesRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	resp, err := portfoliosvc.DescribeChanges(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

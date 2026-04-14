package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/prediction"
)

func RiskDriftHandler(c *gin.Context) {
	var req prediction.DriftRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if len(req.Holdings) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "holdings required"})
		return
	}

	resp, err := prediction.PredictRiskDrift(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

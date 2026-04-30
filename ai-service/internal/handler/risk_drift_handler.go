package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/prediction"
)

func RiskDriftHandler(c *gin.Context) {
	var req prediction.DriftRequest
	if !BindJSONOrRespond(c, &req, "Invalid request") {
		return
	}
	if !RequireNonEmptySlice(c, req.Holdings, "holdings required") {
		return
	}

	resp, err := prediction.PredictRiskDrift(req)
	if err != nil {
		RespondBadRequest(c, "Request failed", err.Error())
		return
	}
	RespondSuccess(c, http.StatusOK, "Risk drift prediction generated", resp)
}

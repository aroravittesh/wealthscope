package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/portfoliosvc"
)

// PortfolioSummarizeHandler handles POST /portfolio/summarize.
func PortfolioSummarizeHandler(c *gin.Context) {
	var req portfoliosvc.SummarizeRequest
	if !BindJSONOrRespond(c, &req, "Invalid JSON") {
		return
	}
	resp, err := portfoliosvc.Summarize(req)
	if err != nil {
		RespondBadRequest(c, "Request failed", err.Error())
		return
	}
	RespondSuccess(c, http.StatusOK, "Portfolio summary generated", resp)
}

// PortfolioChangesHandler handles POST /portfolio/changes.
func PortfolioChangesHandler(c *gin.Context) {
	var req portfoliosvc.ChangesRequest
	if !BindJSONOrRespond(c, &req, "Invalid JSON") {
		return
	}
	resp, err := portfoliosvc.DescribeChanges(req)
	if err != nil {
		RespondBadRequest(c, "Request failed", err.Error())
		return
	}
	RespondSuccess(c, http.StatusOK, "Portfolio changes generated", resp)
}

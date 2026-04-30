package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wealthscope-ai/internal/portfolioexplain"
)

// PortfolioExplainHandler handles POST /portfolio/explain.
func PortfolioExplainHandler(c *gin.Context) {
	var req portfolioexplain.Request
	if !BindJSONOrRespond(c, &req, "Invalid JSON") {
		return
	}

	resp, err := portfolioexplain.Explain(req)
	if err != nil {
		RespondBadRequest(c, "Request failed", err.Error())
		return
	}
	RespondSuccess(c, http.StatusOK, "Portfolio explanation generated", resp)
}

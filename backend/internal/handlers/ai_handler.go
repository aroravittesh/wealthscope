package handlers

import (
	"encoding/json"
	"net/http"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/services"
)

type AIHandler struct {
	Service *services.AIGatewayService
}

func NewAIHandler(service *services.AIGatewayService) *AIHandler {
	return &AIHandler{Service: service}
}

func (h *AIHandler) Recommend(w http.ResponseWriter, r *http.Request) {
	_ = r.Context().Value(middleware.UserIDKey).(string)

	var req models.AIRecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.TopN <= 0 {
		req.TopN = 5
	}
	if req.Risk == "" {
		req.Risk = "medium"
	}
	if req.Horizon == "" {
		req.Horizon = "long"
	}

	res, err := h.Service.Recommend(r.Context(), req)
	if err != nil {
		http.Error(w, "ai recommend failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

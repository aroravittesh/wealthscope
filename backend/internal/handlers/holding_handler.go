package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/services"
)

type HoldingHandler struct {
	Service *services.HoldingService
}

func (h *HoldingHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PortfolioID string  `json:"portfolio_id"`
		Symbol      string  `json:"symbol"`
		AssetType   string  `json:"asset_type"`
		Quantity    float64 `json:"quantity"`
		AvgPrice    float64 `json:"avg_price"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := h.Service.AddHolding(
		userID,
		req.PortfolioID,
		req.Symbol,
		req.AssetType,
		req.Quantity,
		req.AvgPrice,
	)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "holding added/updated",
	})
}

func (h *HoldingHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["portfolio_id"]

	data, err := h.Service.GetHoldings(userID, id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func (h *HoldingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]

	err := h.Service.DeleteHolding(userID, id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "holding deleted",
	})
}

func (h *HoldingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	var req struct {
		Quantity float64 `json:"quantity"`
		AvgPrice float64 `json:"avg_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateHolding(userID, id, req.Quantity, req.AvgPrice); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "holding updated",
	})
}

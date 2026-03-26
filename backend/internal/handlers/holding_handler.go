package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"wealthscope-backend/internal/services"
)

type HoldingHandler struct {
	Service *services.HoldingService
}

func (h *HoldingHandler) Add(w http.ResponseWriter, r *http.Request) {

	var req struct {
		PortfolioID string  `json:"portfolio_id"`
		Symbol      string  `json:"symbol"`
		AssetType   string  `json:"asset_type"`
		Quantity    float64 `json:"quantity"`
		AvgPrice    float64 `json:"avg_price"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := h.Service.AddHolding(
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

	id := mux.Vars(r)["portfolio_id"]

	data, err := h.Service.GetHoldings(id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(data)
}

func (h *HoldingHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	err := h.Service.DeleteHolding(id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "holding deleted",
	})
}
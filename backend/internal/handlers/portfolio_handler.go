package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/services"
)

type PortfolioHandler struct {
	Service *services.PortfolioService
}

func NewPortfolioHandler(service *services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{Service: service}
}

type createRequest struct {
	Name string `json:"name"`
}

func (h *PortfolioHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	p, err := h.Service.Create(userID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(p)
}

func (h *PortfolioHandler) GetUserPortfolios(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	portfolios, err := h.Service.GetUserPortfolios(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(portfolios)
}

func (h *PortfolioHandler) Rename(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := mux.Vars(r)["id"]

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.Service.Rename(userID, id, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PortfolioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := mux.Vars(r)["id"]

	err := h.Service.Delete(userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

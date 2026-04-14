package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
	"wealthscope-backend/internal/services"
)

type ReportingHandler struct {
	PortfolioService *services.PortfolioService
	SnapshotRepo     repository.PortfolioSnapshotRepository
}

func NewReportingHandler(
	portfolioService *services.PortfolioService,
	snapshotRepo repository.PortfolioSnapshotRepository,
) *ReportingHandler {
	return &ReportingHandler{
		PortfolioService: portfolioService,
		SnapshotRepo:     snapshotRepo,
	}
}

// CreatePortfolioSnapshot saves the current analytics summary JSON for a portfolio (owner only).
func (h *ReportingHandler) CreatePortfolioSnapshot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	portfolioID := mux.Vars(r)["id"]

	summary, err := h.PortfolioService.GetPortfolioSummary(userID, portfolioID)
	if err != nil {
		if err.Error() == "unauthorized" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "portfolio not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := json.Marshal(summary)
	if err != nil {
		http.Error(w, "failed to encode snapshot", http.StatusInternalServerError)
		return
	}

	snap := &models.PortfolioSnapshot{
		PortfolioID: portfolioID,
		UserID:      userID,
		SummaryJSON: string(raw),
	}
	if err := h.SnapshotRepo.Create(snap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(snap)
}

// ListPortfolioSnapshots returns saved snapshots for a portfolio (owner only).
func (h *ReportingHandler) ListPortfolioSnapshots(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	portfolioID := mux.Vars(r)["id"]

	// ownership: summary call validates portfolio belongs to user
	if _, err := h.PortfolioService.GetPortfolioSummary(userID, portfolioID); err != nil {
		if err.Error() == "unauthorized" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "portfolio not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := h.SnapshotRepo.ListByPortfolio(portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

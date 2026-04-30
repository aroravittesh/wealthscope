package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

// ComparePortfolioSnapshots returns high-level metric deltas between two snapshots.
// Query params required: from, to (snapshot ids).
func (h *ReportingHandler) ComparePortfolioSnapshots(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	portfolioID := mux.Vars(r)["id"]
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

	fromID := strings.TrimSpace(r.URL.Query().Get("from"))
	toID := strings.TrimSpace(r.URL.Query().Get("to"))
	if fromID == "" || toID == "" {
		http.Error(w, "from and to snapshot ids are required", http.StatusBadRequest)
		return
	}
	if fromID == toID {
		http.Error(w, "from and to must be different snapshot ids", http.StatusBadRequest)
		return
	}

	list, err := h.SnapshotRepo.ListByPortfolio(portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var fromSnap *models.PortfolioSnapshot
	var toSnap *models.PortfolioSnapshot
	for i := range list {
		if list[i].ID == fromID {
			fromSnap = &list[i]
		}
		if list[i].ID == toID {
			toSnap = &list[i]
		}
	}
	if fromSnap == nil || toSnap == nil {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}

	var fromSummary models.PortfolioSummary
	if err := json.Unmarshal([]byte(fromSnap.SummaryJSON), &fromSummary); err != nil {
		http.Error(w, "invalid from snapshot summary", http.StatusInternalServerError)
		return
	}
	var toSummary models.PortfolioSummary
	if err := json.Unmarshal([]byte(toSnap.SummaryJSON), &toSummary); err != nil {
		http.Error(w, "invalid to snapshot summary", http.StatusInternalServerError)
		return
	}

	resp := models.PortfolioSnapshotCompareResponse{
		PortfolioID:          portfolioID,
		FromID:               fromID,
		ToID:                 toID,
		FromAt:               fromSnap.CreatedAt,
		ToAt:                 toSnap.CreatedAt,
		TotalValueDelta:      metricDelta(fromSummary.TotalPortfolioValue, toSummary.TotalPortfolioValue),
		TotalInvestedDelta:   metricDelta(fromSummary.TotalInvested, toSummary.TotalInvested),
		ProfitLossDelta:      metricDelta(fromSummary.TotalProfitLoss, toSummary.TotalProfitLoss),
		DiversificationDelta: metricDelta(fromSummary.DiversificationScore, toSummary.DiversificationScore),
		VolatilityDelta:      metricDelta(fromSummary.VolatilityScore, toSummary.VolatilityScore),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func metricDelta(from float64, to float64) models.SnapshotDelta {
	abs := to - from
	pct := 0.0
	if from != 0 {
		pct = (abs / from) * 100
	}
	return models.SnapshotDelta{
		Absolute: abs,
		Percent:  pct,
	}
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
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
		AllocationDrift:      allocationDrift(fromSummary, toSummary),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetPortfolioSnapshotTrend returns chart-ready metrics for snapshots over time.
// Optional query param: limit (default 20, max 200).
func (h *ReportingHandler) GetPortfolioSnapshotTrend(w http.ResponseWriter, r *http.Request) {
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

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		if n > 200 {
			n = 200
		}
		limit = n
	}

	list, err := h.SnapshotRepo.ListByPortfolio(portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(list) > limit {
		list = list[:limit]
	}
	// Repo returns desc; flip for chronological charting.
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})

	points := make([]models.PortfolioSnapshotTrendPoint, 0, len(list))
	for _, s := range list {
		var summary models.PortfolioSummary
		if err := json.Unmarshal([]byte(s.SummaryJSON), &summary); err != nil {
			http.Error(w, "invalid snapshot summary", http.StatusInternalServerError)
			return
		}
		points = append(points, models.PortfolioSnapshotTrendPoint{
			SnapshotID:          s.ID,
			CreatedAt:           s.CreatedAt,
			TotalPortfolioValue: summary.TotalPortfolioValue,
			TotalInvested:       summary.TotalInvested,
			TotalProfitLoss:     summary.TotalProfitLoss,
			Diversification:     summary.DiversificationScore,
			Volatility:          summary.VolatilityScore,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.PortfolioSnapshotTrendResponse{
		PortfolioID: portfolioID,
		Points:      points,
	})
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

func allocationDrift(from models.PortfolioSummary, to models.PortfolioSummary) []models.AllocationDriftRow {
	fromMap := make(map[string]models.AssetAllocationRow, len(from.AssetAllocation))
	for _, row := range from.AssetAllocation {
		fromMap[row.Symbol] = row
	}
	toMap := make(map[string]models.AssetAllocationRow, len(to.AssetAllocation))
	for _, row := range to.AssetAllocation {
		toMap[row.Symbol] = row
	}

	seen := make(map[string]struct{}, len(fromMap)+len(toMap))
	out := make([]models.AllocationDriftRow, 0, len(fromMap)+len(toMap))

	for symbol, f := range fromMap {
		t, ok := toMap[symbol]
		if !ok {
			t = models.AssetAllocationRow{Symbol: symbol}
		}
		out = append(out, models.AllocationDriftRow{
			Symbol:       symbol,
			FromPercent:  f.Percent,
			ToPercent:    t.Percent,
			DeltaPercent: t.Percent - f.Percent,
			FromValue:    f.Value,
			ToValue:      t.Value,
			DeltaValue:   t.Value - f.Value,
		})
		seen[symbol] = struct{}{}
	}
	for symbol, t := range toMap {
		if _, ok := seen[symbol]; ok {
			continue
		}
		out = append(out, models.AllocationDriftRow{
			Symbol:       symbol,
			FromPercent:  0,
			ToPercent:    t.Percent,
			DeltaPercent: t.Percent,
			FromValue:    0,
			ToValue:      t.Value,
			DeltaValue:   t.Value,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return math.Abs(out[i].DeltaPercent) > math.Abs(out[j].DeltaPercent)
	})
	return out
}

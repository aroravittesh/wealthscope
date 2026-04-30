package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/models"
	"wealthscope-backend/internal/repository"
)

type AdminHandler struct {
	UserRepo     repository.UserRepository
	AssetRepo    repository.AssetRepository
	AuditLogRepo repository.AuditLogRepository
}

func NewAdminHandler(
	userRepo repository.UserRepository,
	assetRepo repository.AssetRepository,
	auditLogRepo repository.AuditLogRepository,
) *AdminHandler {
	return &AdminHandler{UserRepo: userRepo, AssetRepo: assetRepo, AuditLogRepo: auditLogRepo}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserRepo.ListAllPublic()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

type adminUpdateRoleRequest struct {
	Role string `json:"role"`
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	before, err := h.UserRepo.FindByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req adminUpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	role := strings.ToUpper(strings.TrimSpace(req.Role))
	if role != "USER" && role != "ADMIN" {
		http.Error(w, "role must be USER or ADMIN", http.StatusBadRequest)
		return
	}
	if err := h.UserRepo.UpdateRole(id, role); err != nil {
		if err.Error() == "user not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.AuditLogRepo != nil {
		actorUserID, _ := r.Context().Value(middleware.UserIDKey).(string)
		beforeJSON, _ := json.Marshal(map[string]string{"role": before.Role})
		afterJSON, _ := json.Marshal(map[string]string{"role": role})
		_ = h.AuditLogRepo.Create(&models.AuditLog{
			ActorUserID: actorUserID,
			Action:      "USER_ROLE_UPDATED",
			EntityType:  "user",
			EntityID:    id,
			BeforeJSON:  string(beforeJSON),
			AfterJSON:   string(afterJSON),
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := h.AssetRepo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(assets)
}

type adminCreateAssetRequest struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	AssetType string `json:"asset_type"`
}

func (h *AdminHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req adminCreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	at := strings.ToUpper(strings.TrimSpace(req.AssetType))
	if symbol == "" || (at != "STOCK" && at != "CRYPTO" && at != "ETF") {
		http.Error(w, "symbol and asset_type (STOCK|CRYPTO|ETF) required", http.StatusBadRequest)
		return
	}
	a := &models.Asset{Symbol: symbol, Name: strings.TrimSpace(req.Name), AssetType: at}
	if err := h.AssetRepo.Create(a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created, err := h.AssetRepo.FindByID(a.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

type adminUpdateAssetRequest struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	AssetType string `json:"asset_type"`
}

func (h *AdminHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req adminUpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	at := strings.ToUpper(strings.TrimSpace(req.AssetType))
	if symbol == "" || (at != "STOCK" && at != "CRYPTO" && at != "ETF") {
		http.Error(w, "symbol and asset_type (STOCK|CRYPTO|ETF) required", http.StatusBadRequest)
		return
	}
	if err := h.AssetRepo.Update(id, symbol, strings.TrimSpace(req.Name), at); err != nil {
		if err.Error() == "asset not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a, err := h.AssetRepo.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

func (h *AdminHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.AssetRepo.Delete(id); err != nil {
		if err.Error() == "asset not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

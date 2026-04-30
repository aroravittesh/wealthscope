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

func (h *AdminHandler) writeAuditLog(r *http.Request, action string, entityType string, entityID string, before any, after any) {
	if h.AuditLogRepo == nil {
		return
	}
	actorUserID, _ := r.Context().Value(middleware.UserIDKey).(string)
	var beforeText string
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeText = string(b)
		}
	}
	var afterText string
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterText = string(b)
		}
	}
	_ = h.AuditLogRepo.Create(&models.AuditLog{
		ActorUserID: actorUserID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		BeforeJSON:  beforeText,
		AfterJSON:   afterText,
	})
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

	h.writeAuditLog(r, "USER_ROLE_UPDATED", "user", id, map[string]string{"role": before.Role}, map[string]string{"role": role})
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
	h.writeAuditLog(r, "ASSET_CREATED", "asset", created.ID, nil, created)
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
	before, err := h.AssetRepo.FindByID(id)
	if err != nil {
		if err.Error() == "asset not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	h.writeAuditLog(r, "ASSET_UPDATED", "asset", id, before, a)
}

func (h *AdminHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	before, err := h.AssetRepo.FindByID(id)
	if err != nil {
		if err.Error() == "asset not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.AssetRepo.Delete(id); err != nil {
		if err.Error() == "asset not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.writeAuditLog(r, "ASSET_DELETED", "asset", id, before, nil)
}

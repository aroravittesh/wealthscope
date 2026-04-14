package handlers

import (
	"encoding/json"
	"net/http"

	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/repository"
)

type ProfileHandler struct {
	UserRepo repository.UserRepository
}

type profileResponse struct {
	Email          string `json:"email"`
	RiskPreference string `json:"risk_preference"`
}

type updateProfileRequest struct {
	RiskPreference string `json:"risk_preference"`
}

func NewProfileHandler(userRepo repository.UserRepository) *ProfileHandler {
	return &ProfileHandler{UserRepo: userRepo}
}

// GetProfile returns the current user's email and risk preference.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.UserRepo.FindByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	resp := profileResponse{
		Email:          user.Email,
		RiskPreference: user.RiskPreference,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateProfile updates only the risk preference for the current user.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.RiskPreference == "" {
		http.Error(w, "risk_preference is required", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.UpdateRiskPreference(userID, req.RiskPreference); err != nil {
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	user, err := h.UserRepo.FindByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	resp := profileResponse{
		Email:          user.Email,
		RiskPreference: user.RiskPreference,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

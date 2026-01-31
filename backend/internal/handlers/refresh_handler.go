package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"wealthscope-backend/internal/services"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func Refresh(auth *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Find refresh token in DB
		tokenData, err := auth.RefreshTokenRepo.Find(req.RefreshToken)
		if err != nil || tokenData.ExpiresAt.Before(time.Now().UTC()) {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}

		// Update last-used timestamp (inactivity tracking)
		auth.RefreshTokenRepo.UpdateLastUsed(req.RefreshToken, time.Now().UTC())

		// Generate new access token via service
		accessToken, err := auth.RefreshAccessToken(tokenData.UserID)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(refreshResponse{
			AccessToken: accessToken,
		})
	}
}

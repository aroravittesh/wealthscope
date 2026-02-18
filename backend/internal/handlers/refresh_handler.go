package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"wealthscope-backend/internal/services"
	"wealthscope-backend/internal/repository"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}


func Refresh(auth *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Find existing refresh token
		oldToken, err := auth.RefreshTokenRepo.Find(req.RefreshToken)
		if err != nil || oldToken.ExpiresAt.Before(time.Now().UTC()) {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}

		// Rotate: delete old token
		_ = auth.RefreshTokenRepo.Delete(req.RefreshToken)

		// Generate new refresh token
		newRefreshToken, err := services.GenerateNewRefreshToken()
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC()

		newToken := &repository.RefreshToken{
			UserID:     oldToken.UserID,
			Token:      newRefreshToken,
			LastUsedAt: now,
			ExpiresAt:  now.Add(1 * time.Hour),
		}

		if err := auth.RefreshTokenRepo.Create(newToken); err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		// Generate new access token
		accessToken, err := auth.RefreshAccessToken(oldToken.UserID)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(refreshResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
		})
	}
}

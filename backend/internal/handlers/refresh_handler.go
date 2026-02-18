package handlers

import (
	"encoding/json"
	"net/http"

	"wealthscope-backend/internal/services"
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

		accessToken, newRefreshToken, err :=
			auth.RefreshAccessToken(req.RefreshToken)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(refreshResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
		})
	}
}

package handlers

import (
	"encoding/json"
	"net/http"

	"wealthscope-backend/internal/services"
)

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func Logout(auth *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req logoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.RefreshToken == "" {
			http.Error(w, "refresh token required", http.StatusBadRequest)
			return
		}

		// Delete refresh token (invalidate session)
		_ = auth.RefreshTokenRepo.Delete(req.RefreshToken)

		json.NewEncoder(w).Encode(map[string]string{
			"message": "logged out successfully",
		})
	}
}

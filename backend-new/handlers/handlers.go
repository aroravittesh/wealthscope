package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock-backend/models"
	"stock-backend/services"
	"context"
)

// HealthCheckHandler handles GET /health requests
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 Health check request received")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.HealthCheckResponse{
		Status:  "success",
		Message: "Server is running",
	}
	json.NewEncoder(w).Encode(response)
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("❌ 404 Not Found: %s %s", r.Method, r.RequestURI)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	response := models.ErrorResponse{
		Error:   "not_found",
		Message: "Endpoint not found",
		Code:    http.StatusNotFound,
	}

	json.NewEncoder(w).Encode(response)
}

// StockRecommendationHandler handles POST /api/v1/recommend requests
// FUTURE IMPLEMENTATION:
// This handler will:
// 1. Validate the incoming StockRecommendationRequest
// 2. Call the FastAPI ML microservice
// 3. Return the prediction score with confidence
func StockRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔮 Stock recommendation request received")
	w.Header().Set("Content-Type", "application/json")

	// Decode request
	var req models.StockRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request payload",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Call service layer for business logic
	service := services.NewStockService()
	resp, err := service.GetRecommendation(context.Background(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

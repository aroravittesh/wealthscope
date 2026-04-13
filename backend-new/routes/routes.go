package routes

import (
	"net/http"
	"stock-backend/handlers"
	"stock-backend/middleware"

	"github.com/gorilla/mux"
)

// SetupRoutes registers all API routes and middleware on the provided router
// FUTURE: /api/v1/recommend will call FastAPI ML microservice for predictions
func SetupRoutes(router *mux.Router) {
	// Apply global middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.RecoveryMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Stock recommendation endpoint (no business logic here)
	api.HandleFunc("/recommend", handlers.StockRecommendationHandler).Methods("POST")

	// 404 handler for undefined routes
	router.NotFoundHandler = http.HandlerFunc(handlers.NotFoundHandler)
}

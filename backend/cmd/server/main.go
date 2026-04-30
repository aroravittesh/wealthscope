package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"wealthscope-backend/internal/db"
	"wealthscope-backend/internal/handlers"
	"wealthscope-backend/internal/market"
	"wealthscope-backend/internal/middleware"
	"wealthscope-backend/internal/repository"
	"wealthscope-backend/internal/services"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")

	database := db.Connect()
	defer database.Close()

	// repositories
	userRepo := repository.NewUserRepository(database)
	portfolioRepo := repository.NewPortfolioRepository(database)
	holdingRepo := repository.NewHoldingRepository(database)
	assetRepo := repository.NewAssetRepository(database)
	snapshotRepo := repository.NewPortfolioSnapshotRepository(database)
	auditLogRepo := repository.NewAuditLogRepository(database)

	portfolioService := &services.PortfolioService{
		PortfolioRepo: portfolioRepo,
		HoldingRepo:   holdingRepo,
		Prices:        market.NewDefaultProvider(),
	}
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)

	// ✅ holdings
	holdingService := &services.HoldingService{
		Repo:          holdingRepo,
		PortfolioRepo: portfolioRepo,
	}
	holdingHandler := &handlers.HoldingHandler{Service: holdingService}

	refreshTokenRepo := repository.NewRefreshTokenRepository(database)

	// services
	authService := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
	}

	// handlers
	authHandler := handlers.NewAuthHandler(authService)
	profileHandler := handlers.NewProfileHandler(userRepo)
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:8000"
	}
	aiGatewayService := services.NewAIGatewayService(aiServiceURL)
	aiHandler := handlers.NewAIHandler(aiGatewayService)

	adminHandler := handlers.NewAdminHandler(userRepo, assetRepo, auditLogRepo)
	reportingHandler := handlers.NewReportingHandler(portfolioService, snapshotRepo)

	adminOnly := func(h http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(middleware.RequireRole("ADMIN")(h))
	}

	router := mux.NewRouter()

	// Subrouter for API routes
	api := router.PathPrefix("/api").Subrouter()

	// health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WealthScope backend connected to Supabase"))
	}).Methods("GET")

	// auth routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", handlers.Refresh(authService)).Methods("POST")
	api.HandleFunc("/auth/logout", handlers.Logout(authService)).Methods("POST")

	api.Handle(
		"/auth/change-password",
		middleware.AuthMiddleware(
			handlers.ChangePassword(authService),
		),
	).Methods("POST")

	// profile routes
	api.Handle(
		"/auth/profile",
		middleware.AuthMiddleware(http.HandlerFunc(profileHandler.GetProfile)),
	).Methods("GET")

	api.Handle(
		"/auth/profile",
		middleware.AuthMiddleware(http.HandlerFunc(profileHandler.UpdateProfile)),
	).Methods("PUT")

	// portfolio routes
	api.Handle(
		"/portfolios",
		middleware.AuthMiddleware(http.HandlerFunc(portfolioHandler.Create)),
	).Methods("POST")

	api.Handle(
		"/portfolios",
		middleware.AuthMiddleware(http.HandlerFunc(portfolioHandler.GetUserPortfolios)),
	).Methods("GET")

	api.Handle(
		"/portfolios/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(portfolioHandler.Rename)),
	).Methods("PUT")

	api.Handle(
		"/portfolios/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(portfolioHandler.Delete)),
	).Methods("DELETE")

	api.Handle(
		"/portfolios/{id}/summary",
		middleware.AuthMiddleware(http.HandlerFunc(portfolioHandler.GetSummary)),
	).Methods("GET")

	api.Handle(
		"/portfolios/{id}/snapshots",
		middleware.AuthMiddleware(http.HandlerFunc(reportingHandler.CreatePortfolioSnapshot)),
	).Methods("POST")

	api.Handle(
		"/portfolios/{id}/snapshots",
		middleware.AuthMiddleware(http.HandlerFunc(reportingHandler.ListPortfolioSnapshots)),
	).Methods("GET")

	// admin (JWT role ADMIN)
	api.Handle("/admin/audit-logs", adminOnly(adminHandler.ListAuditLogs)).Methods("GET")
	api.Handle("/admin/users", adminOnly(adminHandler.ListUsers)).Methods("GET")
	api.Handle("/admin/users/{id}/role", adminOnly(adminHandler.UpdateUserRole)).Methods("PATCH")
	api.Handle("/admin/assets", adminOnly(adminHandler.ListAssets)).Methods("GET")
	api.Handle("/admin/assets", adminOnly(adminHandler.CreateAsset)).Methods("POST")
	api.Handle("/admin/assets/{id}", adminOnly(adminHandler.UpdateAsset)).Methods("PUT")
	api.Handle("/admin/assets/{id}", adminOnly(adminHandler.DeleteAsset)).Methods("DELETE")

	// ✅ holdings routes
	api.Handle(
		"/holdings",
		middleware.AuthMiddleware(http.HandlerFunc(holdingHandler.Add)),
	).Methods("POST")

	api.Handle(
		"/holdings/{portfolio_id}",
		middleware.AuthMiddleware(http.HandlerFunc(holdingHandler.Get)),
	).Methods("GET")

	api.Handle(
		"/holdings/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(holdingHandler.Delete)),
	).Methods("DELETE")

	api.Handle(
		"/holdings/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(holdingHandler.Update)),
	).Methods("PUT")

	// AI recommendation gateway routes (Python ML service)
	api.Handle(
		"/ai/recommend",
		middleware.AuthMiddleware(http.HandlerFunc(aiHandler.Recommend)),
	).Methods("POST")

	// CORS middleware
	withCORS := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowedOrigins := map[string]bool{
				"http://localhost:4200":          true,
				"https://aurex-sable.vercel.app": true,
			}

			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	log.Println("WealthScope server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(router)))
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"wealthscope-backend/internal/db"
	"wealthscope-backend/internal/handlers"
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
	portfolioService := &services.PortfolioService{
		PortfolioRepo: portfolioRepo,
	}
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)

	refreshTokenRepo := repository.NewRefreshTokenRepository(database)

	// services (DEPENDENCIES WIRED CORRECTLY)
	authService := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
	}

	// handlers
	authHandler := handlers.NewAuthHandler(authService)
	profileHandler := handlers.NewProfileHandler(userRepo)

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

	// basic CORS middleware for local Angular dev
	withCORS := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
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

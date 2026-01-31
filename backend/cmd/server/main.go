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
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)

	// services (DEPENDENCIES WIRED CORRECTLY)
	authService := &services.AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
	}

	// handlers
	authHandler := handlers.NewAuthHandler(authService)

	router := mux.NewRouter()

	// health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WealthScope backend connected to Supabase"))
	}).Methods("GET")

	// auth routes
	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/auth/refresh", handlers.Refresh(authService)).Methods("POST")
	router.HandleFunc("/auth/logout", handlers.Logout(authService)).Methods("POST")

	router.Handle(
		"/auth/change-password",
		middleware.AuthMiddleware(
			handlers.ChangePassword(authService),
		),
	).Methods("POST")
	


	log.Println("WealthScope server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"wealthscope-backend/internal/db"
	"wealthscope-backend/internal/handlers"
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

	// services
	authService := &services.AuthService{
		UserRepo: userRepo,
	}

	// handlers
	authHandler := handlers.NewAuthHandler(authService)

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WealthScope backend connected to Supabase"))
	}).Methods("GET")

	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	log.Println("WealthScope server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

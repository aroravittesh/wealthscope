package main

import (
	"log"
	"net/http"
	"os"

	"stock-backend/routes"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Setup routes
	routes.SetupRoutes(router)

	// Read port from env (IMPORTANT FIX)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running at http://localhost:" + port)

	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}

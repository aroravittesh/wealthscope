package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Host string
	Port string
	Env  string
	// DB placeholder for future SQLite connection
	// DB *sql.DB // FUTURE: Add DB connection here
}

// LoadConfig loads configuration from environment or returns defaults
func LoadConfig() *Config {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// FUTURE: Initialize DB connection here
	// db, err := sql.Open("sqlite3", "./foo.db")
	// if err != nil { ... }

	return &Config{
		Host: host,
		Port: port,
		Env:  env,
		// DB: db, // FUTURE: Add DB connection here
	}
}

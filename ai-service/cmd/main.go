package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"wealthscope-ai/internal/handler"
)

func main() {

	godotenv.Load()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "AI Service running",
		})
	})

	router.POST("/chat", handler.ChatHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	router.Run(":" + port)
}
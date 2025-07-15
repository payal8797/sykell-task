package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/payal8797/sykell-task/backend/db"
	"github.com/payal8797/sykell-task/backend/handlers"
)

func main() {
	// Initialize DB connection
	db.InitDB()

	// Set Gin mode from environment variable (optional but useful)
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// Create new Gin router with default middleware (logger, recovery)
	r := gin.Default()

	// Health check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Register all routes in a group
	urlRoutes := r.Group("/urls")
	{
		urlRoutes.POST("", handlers.PostURL)                  // Create and crawl a new URL
		urlRoutes.GET("", handlers.GetAllURLs)                // Get all URLs
		urlRoutes.GET("/:id", handlers.GetURLByID)            // Get details of a single URL
		urlRoutes.POST("/:id/reanalyze", handlers.ReanalyzeURL) // Reanalyze a specific URL
		urlRoutes.DELETE("/:id", handlers.DeleteURL)          // Delete a URL
	}

	// Start the server
	log.Println("🚀 Server running at http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

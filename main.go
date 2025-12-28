package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/app"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	database.ConnectDatabase(cfg)

	// Initialize Gin router
	r := gin.Default()

	// Setup Routes
	app.SetupRoutes(r)

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

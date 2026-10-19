package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/app"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/realtime"
)

func main() {
	resetPtr := flag.Bool("reset", false, "Reset the database (drop all tables)")
	flag.Parse()

	cfg := config.LoadConfig()

	database.ConnectDatabase(cfg)

	if *resetPtr {
		log.Println("Resetting database...")
		database.ResetDatabase()
		log.Println("Database reset complete.")
		return
	}

	r := gin.Default()

	hub := realtime.NewHub()
	// each room in hub runs its own loop

	app.SetupRoutes(r, hub)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

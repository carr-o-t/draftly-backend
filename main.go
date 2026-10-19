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
	// command line flags
	resetPtr := flag.Bool("reset", false, "Reset the database (drop all tables)")
	flag.Parse()

	// load config
	cfg := config.LoadConfig()

	// connect to the database
	database.ConnectDatabase(cfg)

	if *resetPtr {
		log.Println("Resetting database...")
		database.ResetDatabase()
		log.Println("Database reset complete.")
		return
	}

	// set up the gin router
	r := gin.Default()

	// set up the realtime hub
	hub := realtime.NewHub()
	// the hub has no run loop of its own, each room runs its own loop

	// register all routes
	app.SetupRoutes(r, hub)

	// start the server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

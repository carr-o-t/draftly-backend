package main

import (
	"flag"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/app"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/realtime"
)

// splitAndTrim turns "a,b" into ["a", "b"] and drops empty parts
func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

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

	// on railway we sit behind their proxy, so we trust the forwarded
	// headers and read the real client ip from X-Forwarded-For. locally we
	// trust no proxy. set TRUSTED_PROXIES="0.0.0.0/0" on the platform.
	if cfg.TrustedProxies != "" {
		if err := r.SetTrustedProxies(splitAndTrim(cfg.TrustedProxies)); err != nil {
			log.Fatalf("Failed to set trusted proxies: %v", err)
		}
	} else {
		// nil means trust nobody
		r.SetTrustedProxies(nil)
	}

	hub := realtime.NewHub()
	// each room in hub runs its own loop

	app.SetupRoutes(r, hub)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

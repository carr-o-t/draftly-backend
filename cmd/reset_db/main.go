package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

func main() {
	// Load .env relative to this file location or root
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	log.Println("⚠️  WARNING: This will DROP ALL TABLES and DATA. Press Ctrl+C to cancel in 5 seconds...")
	// time.Sleep(5 * time.Second) // Uncomment for safety in prod scripts, removed for dev speed

	log.Println("Dropping tables...")

	// Drop in reverse order of dependency
	// Permissions/Shares depend on Documents/Users
	database.DB.Migrator().DropTable(&models.AccessRequest{})
	database.DB.Migrator().DropTable(&models.DocumentPermission{})
	database.DB.Migrator().DropTable(&models.DocumentShare{})
	database.DB.Migrator().DropTable(&models.Version{})
	database.DB.Migrator().DropTable(&models.IPUsage{})
	database.DB.Migrator().DropTable(&models.Document{})
	database.DB.Migrator().DropTable(&models.User{})

	log.Println("Running AutoMigrate...")
	// Re-run migration logic (calls AutoMigrate internally in ConnectDatabase, but we can be explicit or just rely on it if we reconnect, 
	// actually ConnectDatabase CALLS AutoMigrate. 
	// But since we just dropped them after connection, we need to call it again.
	
	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.DocumentShare{},
		&models.Version{},
		&models.IPUsage{},
		&models.AccessRequest{},
		&models.DocumentPermission{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Database reset successfully!")
}

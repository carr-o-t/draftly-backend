package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

func main() {
	// load .env, path is relative to this file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: Error loading .env file")
	}

	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	log.Println("⚠️  WARNING: This will DROP ALL TABLES and DATA. Press Ctrl+C to cancel in 5 seconds...")
	// time.Sleep(5 * time.Second) // uncomment if you want the safety delay

	log.Println("Dropping tables...")

	// drop children before parents
	// permissions and shares depend on documents and users
	database.DB.Migrator().DropTable(&models.AccessRequest{})
	database.DB.Migrator().DropTable(&models.DocumentPermission{})
	database.DB.Migrator().DropTable(&models.DocumentShare{})
	database.DB.Migrator().DropTable(&models.Version{})
	database.DB.Migrator().DropTable(&models.IPUsage{})
	database.DB.Migrator().DropTable(&models.Document{})
	database.DB.Migrator().DropTable(&models.User{})

	log.Println("Running AutoMigrate...")
	// conndatabase already migrates, but we dropped the tables after connecting,
	// so we have to run it again here
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

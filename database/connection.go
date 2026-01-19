package database

import (
	"log"

	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(cfg *config.Config) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set in environment variables")
	}

	// Use PreferSimpleProtocol to disable implicit prepared statement usage by the driver
	// This is crucial for Supabase Transaction Pooler (port 6543)
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false, // disable GORM's caching
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate
	err = database.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.DocumentShare{}, // Keep for now or deprecate? MVP says replace.
		&models.Version{},
		&models.IPUsage{},
		&models.AccessRequest{},
		&models.DocumentPermission{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	DB = database
	log.Println("Database connection established and migrations applied")
}

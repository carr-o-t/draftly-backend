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

	// PreferSimpleProtocol turns off prepared statements in the driver.
	// this is needed for the supabase transaction pooler on port 6543.
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false, // no gorm statement caching
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// create or update tables
	err = database.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.DocumentShare{}, // still here for now, may be removed later
		&models.Version{},
		&models.IPUsage{},
		&models.AccessRequest{},
		&models.DocumentPermission{},
		&models.DocumentUpdate{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	DB = database
	log.Println("Database connection established and migrations applied")
}

func ResetDatabase() {
	if DB == nil {
		log.Fatal("Database not connected")
	}

	// drop tables children first, this order matters because of foreign keys:
	// documentupdate -> document
	// documentpermission -> document / user
	// accessrequest -> document / user
	// documentshare -> document
	// version -> document
	// document -> user
	// ipusage
	// user
	err := DB.Migrator().DropTable(
		&models.DocumentUpdate{},
		&models.DocumentPermission{},
		&models.AccessRequest{},
		&models.DocumentShare{},
		&models.Version{},
		&models.Document{},
		&models.IPUsage{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}
	log.Println("All tables dropped.")

	// create the tables again
	err = DB.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.DocumentShare{},
		&models.Version{},
		&models.IPUsage{},
		&models.AccessRequest{},
		&models.DocumentPermission{},
		&models.DocumentUpdate{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database re-migrated successfully.")
}

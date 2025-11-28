package dbops

import (
	"log"

	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/supabase-community/supabase-go"
)

var Client *supabase.Client

func ConnectDatabase(cfg *config.Config) {
	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseKey, nil)
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	Client = client
	log.Println("Supabase client initialized")
}

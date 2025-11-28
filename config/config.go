package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL string
	SupabaseKey string
	ServerPort  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, relying on environment variables")
	}

	return &Config{
		SupabaseURL: getEnv("SUPABASE_PROJECT_URL", ""),
		SupabaseKey: getEnv("SUPABASE_API_KEY", ""),
		ServerPort:  getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

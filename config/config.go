package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL    string
	SupabaseKey    string
	ServerPort     string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
	// comma separated list of proxies we trust to set the client ip
	TrustedProxies string
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
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key-change-me"),
		// only this origin is allowed to open a websocket
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		// railway puts us behind its own proxy, so trusting all is fine there.
		// locally there is no proxy, so we trust none.
		TrustedProxies: getEnv("TRUSTED_PROXIES", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

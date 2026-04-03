package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey  string
	NewsAPIKey    string
	Port          string
	GinMode       string
	AllowedOrigin string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}

	return &Config{
		GeminiAPIKey:  getEnv("GOOGLE_GENERATIVE_AI_API_KEY", ""),
		NewsAPIKey:    getEnv("NEWS_API_KEY", ""),
		Port:          getEnv("PORT", "8080"),
		GinMode:       getEnv("GIN_MODE", "debug"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

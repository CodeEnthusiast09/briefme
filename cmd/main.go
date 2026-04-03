package main

import (
	"fmt"
	"log"

	"github.com/CodeEnthusiast09/briefme-backend/internal/config"
	"github.com/CodeEnthusiast09/briefme-backend/internal/handlers"
	"github.com/CodeEnthusiast09/briefme-backend/internal/router"
	"github.com/CodeEnthusiast09/briefme-backend/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.GeminiAPIKey == "" {
		log.Fatal("GOOGLE_GENERATIVE_AI_API_KEY is required but not set")
	}
	if cfg.NewsAPIKey == "" {
		log.Fatal("NEWS_API_KEY is required but not set")
	}

	gin.SetMode(cfg.GinMode)

	geminiService, err := services.NewGeminiService(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("failed to initialize Gemini service: %v", err)
	}

	newsService := services.NewNewsAPIService(cfg.NewsAPIKey)
	scraperService := services.NewScraperService()

	summarizeHandler := handlers.NewSummarizeHandler(geminiService, newsService, scraperService)
	chatHandler := handlers.NewChatHandler(geminiService)

	r := router.Setup(summarizeHandler, chatHandler, cfg.AllowedOrigin)

	// Step 7: Start the server.
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on port %s", cfg.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

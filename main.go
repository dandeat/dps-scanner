package main

import (
	"log"
	"dps-scanner-gateout/config"
	"dps-scanner-gateout/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the Gin router
	r := gin.Default()

	// Set up all application routes
	handlers.SetupRoutes(r, cfg)

	// Start the server
	log.Printf("Server running at http://localhost:%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

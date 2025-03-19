package main

import (
	"log"
	"strconv"

	"github.com/lgm8-measurements-service/api/routes"
	"github.com/lgm8-measurements-service/config"
	"github.com/lgm8-measurements-service/internal/auth"
	"github.com/lgm8-measurements-service/internal/db"
	"github.com/lgm8-measurements-service/internal/httpclient"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: [%s]", err)
	}

	// Initialize database connection
	if err := db.Connect(&cfg.DB); err != nil {
		log.Fatalf("Error connecting to the database: [%s]", err)
	}

	// Initialize intra-microservices http client
	httpClient := httpclient.NewHTTPClient(cfg.NGINX.BaseURL)

	// Initialize JWKS manager
	jwksManager := auth.NewJWKSManager(httpClient, cfg.Auth.JWKS)
	if err := jwksManager.FetchJWKS(); err != nil {
		log.Fatalf("Failed to fetch initial JWKS: [%s]", err)
	}

	// Router setup
	r := routes.SetupRouter()

	// Server startup
	port := strconv.Itoa(cfg.Server.Port)
	log.Printf("Server started on port: [%s]", port)
	r.Run(":" + port)
}

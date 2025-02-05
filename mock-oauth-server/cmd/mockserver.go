package main

import (
	"log"
	"mock-oauth-server/config"
	"mock-oauth-server/internal/app"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create App: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

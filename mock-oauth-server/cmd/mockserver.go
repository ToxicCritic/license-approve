// cmd/mockserver.go
package main

import (
	"log"
	"mock-authserver/config"
	"mock-authserver/internal/app"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[MOCK] Failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("[MOCK] Failed to init mock app: %v", err)
	}

	// Пути к сертификатам
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("[MOCK] Unable to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	certFile := filepath.Join(exeDir, "certs", "mock-oauth.crt")
	keyFile := filepath.Join(exeDir, "certs", "mock-oauth.key")

	log.Printf("[MOCK] Starting mock OAuth2.0 server on %s (TLS)", cfg.HTTPAddr)
	err = http.ListenAndServeTLS(cfg.HTTPAddr, certFile, keyFile, application.Router)
	if err != nil {
		log.Fatalf("[MOCK] Failed to start server: %v", err)
	}
}

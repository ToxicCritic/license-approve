// mock-oauth-server/cmd/main.go

package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"mock-oauth-server/internal/handlers"
	"mock-oauth-server/internal/models"
)

// Config represents the mock-oauth configuration loaded from JSON
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
}

var configData Config

func main() {
	loadConfig()

	// Init an in-memory store with a default user if you like
	models.InitStore() // optional

	// Setup default HTTP transport to skip TLS verification (for local testing)
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	router := mux.NewRouter()

	// OAuth 2.0 endpoints
	router.HandleFunc("/authorize", handlers.AuthorizeHandler).Methods("GET", "POST")
	router.HandleFunc("/token", handlers.TokenHandler).Methods("POST")
	router.HandleFunc("/userinfo", handlers.UserInfoHandler).Methods("GET")

	// Locate certificate and key
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("[MOCK-OAUTH] Unable to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	certFile := filepath.Join(exeDir, "certs", "mock-oauth.crt")
	keyFile := filepath.Join(exeDir, "certs", "mock-oauth.key")

	// Check if files exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("[MOCK-OAUTH] SSL certificate not found at: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("[MOCK-OAUTH] SSL key not found at: %s", keyFile)
	}

	addr := ":8081"
	log.Printf("[MOCK-OAUTH] Starting mock OAuth2.0 server on %s (TLS)", addr)
	if err := http.ListenAndServeTLS(addr, certFile, keyFile, router); err != nil {
		log.Fatalf("[MOCK-OAUTH] Failed to start server: %v", err)
	}
}

func loadConfig() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("[MOCK-OAUTH] Unable to determine executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")

	log.Printf("[MOCK-OAUTH] Loading config from: %s", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("[MOCK-OAUTH] Could not open config.json: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&configData); err != nil {
		log.Fatalf("[MOCK-OAUTH] Could not decode config.json: %v", err)
	}
	log.Println("[MOCK-OAUTH] Configuration loaded successfully")

	// Pass config to handlers package if needed
	handlers.SetConfig(handlers.Config{
		ClientID:     configData.ClientID,
		ClientSecret: configData.ClientSecret,
		RedirectURI:  configData.RedirectURI,
	})
}

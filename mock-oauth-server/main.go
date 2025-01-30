// mock-oauth-server/main.go

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
}

var configData Config

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config.json: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configData); err != nil {
		log.Fatalf("Failed to decode config.json: %v", err)
	}
}

func main() {
	loadConfig()

	r := mux.NewRouter()

	r.HandleFunc("/authorize", authorizeHandler).Methods("GET", "POST")
	r.HandleFunc("/token", tokenHandler).Methods("POST")
	r.HandleFunc("/userinfo", userInfoHandler).Methods("GET")

	addr := ":8081"
	log.Printf("Mock OAuth2.0 server started on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

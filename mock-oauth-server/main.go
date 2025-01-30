package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Не удалось определить путь к исполняемому файлу: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")
	log.Printf("Загрузка конфигурации из: %s", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Не удалось открыть config.json: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configData); err != nil {
		log.Fatalf("Не удалось декодировать config.json: %v", err)
	}

	log.Println("Конфигурация загружена успешно")
}

func main() {
	loadConfig()

	r := mux.NewRouter()

	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // отключить проверку сертификатов
		},
	}

	// OAuth 2.0 Endpoints
	r.HandleFunc("/authorize", authorizeHandler).Methods("GET", "POST")
	r.HandleFunc("/token", tokenHandler).Methods("POST")
	r.HandleFunc("/userinfo", userInfoHandler).Methods("GET")

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Не удалось определить путь к исполняемому файлу: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	certFile := filepath.Join(exeDir, "certs", "mock-oauth.crt")
	keyFile := filepath.Join(exeDir, "certs", "mock-oauth.key")

	// Проверка существования файлов сертификата и ключа
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificate not found at path: %s", certFile)
	} else if err != nil {
		log.Fatalf("Error checking certificate file: %v", err)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Key not found at path: %s", keyFile)
	} else if err != nil {
		log.Fatalf("Error checking key file: %v", err)
	}

	// Запуск HTTPS-сервера
	addr := ":8081"
	log.Printf("Mock OAuth2.0 server started on %s with TLS", addr)
	if err := http.ListenAndServeTLS(addr, certFile, keyFile, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

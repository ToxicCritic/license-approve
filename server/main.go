// server/main.go

package main

import (
	"LicenseApp/server/pkg/auth"
	"LicenseApp/server/pkg/db"
	"LicenseApp/server/pkg/security"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения
	err := loadEnv()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Инициализация OAuth2 конфигурации
	auth.InitOAuthConfig()

	// Инициализация хранилища сессий
	auth.SetupSessionStore()

	// Инициализация базы данных
	db.Init()
	db.Migrate()

	// Настройка безопасности (например, загрузка ключей для JWT или других механизмов)
	configPath := filepath.Join("server", "config")
	privateKeyFile := filepath.Join(configPath, "keys", "private_key.pem")
	publicKeyFile := filepath.Join(configPath, "keys", "public_key.pem")

	err = security.LoadKeys(privateKeyFile, publicKeyFile)
	if err != nil {
		log.Fatalf("Error loading security keys: %v", err)
	}

	// Создание маршрутизатора
	router := mux.NewRouter()

	// Маршруты аутентификации
	router.HandleFunc("/auth/login", auth.LoginHandler).Methods("GET")
	router.HandleFunc("/oauth-cb", auth.CallbackHandler).Methods("GET")
	router.HandleFunc("/auth/logout", auth.LogoutHandler).Methods("GET")

	// Применение middleware к административным маршрутам
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(auth.AuthMiddleware())
	adminRouter.HandleFunc("/license-requests", db.GetLicenseRequestsHandler).Methods("GET")
	adminRouter.HandleFunc("/approve-license", db.ApproveLicenseRequestHandler).Methods("POST")
	adminRouter.HandleFunc("/reject-license", db.RejectLicenseRequestHandler).Methods("POST")

	// Открытые маршруты (не требуют аутентификации)
	router.HandleFunc("/api/check-license", db.CheckLicenseHandler).Methods("GET")
	router.HandleFunc("/api/create-license-request", db.CreateLicenseRequestHandler).Methods("POST")

	// Настройка путей к сертификатам и ключам для HTTPS
	certFile := filepath.Join(configPath, "certs", "server.crt")
	keyFile := filepath.Join(configPath, "certs", "server.key")
	log.Println("Certificate file path:", certFile)
	log.Println("Key file path:", keyFile)

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

	// Запуск HTTPS-сервера с использованием маршрутизатора
	log.Println("HTTPS server started on port 8443...")
	err = http.ListenAndServeTLS(":8443", certFile, keyFile, router)
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
	}
}

// loadEnv загружает переменные окружения из файла .env
func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using environment variables.")
	}
	return nil
}

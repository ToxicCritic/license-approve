// server/main.go

package main

import (
	"LicenseApp/server/config"
	"LicenseApp/server/pkg/auth"
	"LicenseApp/server/pkg/db"
	"LicenseApp/server/pkg/security"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Инициализация OAuth2 конфигурации
	auth.InitOAuthConfig(cfg)

	// Инициализация хранилища сессий
	auth.SetupSessionStore(cfg)

	// Инициализация базы данных
	db.Init()
	db.Migrate()

	// Настройка безопасности (например, загрузка ключей для JWT или других механизмов)
	err = security.LoadKeys(cfg.PrivateKeyPath, cfg.PublicKeyPath)
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
	certFile := cfg.CertFile
	keyFile := cfg.KeyFile
	log.Println("Certificate filepath:", certFile)
	log.Println("Key filepath:", keyFile)

	// Проверка существования файлов сертификата и ключа
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificate not found in directory: %s", certFile)
	} else if err != nil {
		log.Fatalf("Error checking certificate file: %v", err)
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Key not found in directory: %s", keyFile)
	} else if err != nil {
		log.Fatalf("Error checking key: %v", err)
	}

	// Запуск HTTPS-сервера с использованием маршрутизатора
	log.Println("HTTPS server running on port :8443...")
	err = http.ListenAndServeTLS(":8443", certFile, keyFile, router)
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
	}
}

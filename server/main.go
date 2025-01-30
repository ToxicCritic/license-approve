// server/main.go

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"example.com/licence-approval/server/config"
	"example.com/licence-approval/server/pkg/auth"
	"example.com/licence-approval/server/pkg/db"
	"example.com/licence-approval/server/pkg/security"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	// Загрузка конфигурации
	cfg, err := loadConfig()

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
	log.Println("Certificate file path:", certFile)
	log.Println("Key file path:", keyFile)

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

// loadConfig загружает конфигурацию с использованием viper
func loadConfig() (*config.Config, error) {
	viper.SetConfigName(".env") // имя файла конфигурации (без расширения)
	viper.SetConfigType("env")  // тип файла конфигурации
	viper.AddConfigPath(".")    // путь к файлу конфигурации

	// Загрузка переменных окружения
	viper.AutomaticEnv()

	// Чтение конфигурационного файла
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found. Using environment variables.")
	}

	// Считывание конфигурации в структуру
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Проверка обязательных переменных
	if cfg.OAuthClientID == "" || cfg.OAuthClientSecret == "" || cfg.OAuthRedirectURL == "" ||
		cfg.OAuthAuthURL == "" || cfg.OAuthTokenURL == "" || cfg.SessionSecret == "" ||
		cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" ||
		cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("missing required configuration parameters")
	}

	return &cfg, nil
}

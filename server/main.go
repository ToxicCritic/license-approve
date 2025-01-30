package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"example.com/licence-approval/server/config"
	"example.com/licence-approval/server/pkg/auth"
	"example.com/licence-approval/server/pkg/db"
	"example.com/licence-approval/server/pkg/security"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	// Загрузка конфигурации (через viper, .env лежит рядом с бинарником)
	cfg, err := loadConfigSameDirAsBinary()

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

	// Настройка безопасности (например, загрузка ключей для RSA-ключей)
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

	// Настройка путей к сертификатам
	certFile := cfg.CertFile
	keyFile := cfg.KeyFile
	log.Println("Certificate file path:", certFile)
	log.Println("Key file path:", keyFile)

	// Проверка сертификата и ключа
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

	// Запуск HTTPS-сервера
	log.Println("HTTPS server started on port 8443...")

	err = http.ListenAndServeTLS(":8443", certFile, keyFile, router)
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
	}
}

// Ищет файл ".env" рядом с бинарником (server),
// а если его нет — использует переменные окружения.
func loadConfigSameDirAsBinary() (*config.Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Формируем путь к .env в той же папке
	envPath := filepath.Join(exeDir, ".env")

	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found in %s (using environment vars). Error: %v\n", exeDir, err)
	}

	// Разрешаем переопределение переменных окружением
	viper.AutomaticEnv()

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if cfg.OAuthClientID == "" || cfg.OAuthClientSecret == "" || cfg.OAuthRedirectURL == "" ||
		cfg.OAuthAuthURL == "" || cfg.OAuthTokenURL == "" || cfg.SessionSecret == "" ||
		cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" ||
		cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("missing required configuration parameters")
	}

	return &cfg, nil
}

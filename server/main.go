// server/main.go

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
	cfg, err := loadConfigSameDirAsBinary()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Настраиваем OAuth2
	auth.InitOAuthConfig(cfg)
	auth.SetupSessionStore(cfg)

	// DB init
	db.Init()
	db.Migrate()

	// Загрузка ключей (если нужно для лицензий)
	err = security.LoadKeys(cfg.PrivateKeyPath, cfg.PublicKeyPath)
	if err != nil {
		log.Fatalf("Error loading security keys: %v", err)
	}

	router := mux.NewRouter()

	// Роуты авторизации
	router.HandleFunc("/auth/login", auth.LoginHandler).Methods("GET")
	router.HandleFunc("/oauth-cb", auth.CallbackHandler).Methods("GET")
	router.HandleFunc("/auth/logout", auth.LogoutHandler).Methods("GET")

	// Админские маршруты
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(auth.AuthMiddleware())
	adminRouter.HandleFunc("/license-requests", db.GetLicenseRequestsHandler).Methods("GET")
	adminRouter.HandleFunc("/approve-license", db.ApproveLicenseRequestHandler).Methods("POST")
	adminRouter.HandleFunc("/reject-license", db.RejectLicenseRequestHandler).Methods("POST")

	// Открытые маршруты
	router.HandleFunc("/api/check-license", db.CheckLicenseHandler).Methods("GET")
	router.HandleFunc("/api/create-license-request", db.CreateLicenseRequestHandler).Methods("POST")

	log.Println("Certificate:", cfg.CertFile)
	log.Println("KeyFile:", cfg.KeyFile)
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		log.Fatalf("No cert file %s", cfg.CertFile)
	}
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		log.Fatalf("No key file %s", cfg.KeyFile)
	}

	log.Println("Server on :8443 ...")
	err = http.ListenAndServeTLS(":8443", cfg.CertFile, cfg.KeyFile, router)
	if err != nil {
		log.Fatalf("ListenAndServeTLS error: %v", err)
	}
}

// Загружает .env рядом с бинарником
func loadConfigSameDirAsBinary() (*config.Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env in %s, using environment: %v\n", exeDir, err)
	}
	viper.AutomaticEnv()

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}
	// Check required
	if cfg.OAuthClientID == "" || cfg.OAuthClientSecret == "" ||
		cfg.OAuthRedirectURL == "" || cfg.OAuthAuthURL == "" || cfg.OAuthTokenURL == "" ||
		cfg.SessionSecret == "" || cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("missing required config fields")
	}
	return &cfg, nil
}

package main

import (
	"log"
	"net/http"
	"os"

	"server/config"
	"server/pkg/auth"
	"server/pkg/db"
	"server/pkg/security"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Настраиваем OAuth2
	auth.InitOAuthConfig(cfg)
	auth.SetupSessionStore(cfg)

	// DB init
	db.Init(cfg)
	db.Migrate()

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

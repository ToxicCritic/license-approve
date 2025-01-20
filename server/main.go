// server/main.go
package main

import (
	"LicenseApp/server/pkg/db"
	"LicenseApp/server/pkg/handlers"
	"LicenseApp/server/pkg/security"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Инициализация базы данных
	db.Init()
	db.Migrate()

	// Загрузка ключей для подписания лицензий
	configPath := filepath.Join("config")
	privateKeyFile := filepath.Join(configPath, "keys", "private_key.pem")
	publicKeyFile := filepath.Join(configPath, "keys", "public_key.pem")

	err := security.LoadKeys(privateKeyFile, publicKeyFile)
	if err != nil {
		log.Fatalf("Error loading security keys: %v", err)
	}

	// Регистрация HTTP-обработчиков
	http.HandleFunc("/admin/license-requests", handlers.GetLicenseRequestsHandler)
	http.HandleFunc("/admin/approve-license", handlers.ApproveLicenseRequestHandler)
	http.HandleFunc("/admin/reject-license", handlers.RejectLicenseRequestHandler)
	http.HandleFunc("/api/check-license", handlers.CheckLicenseHandler)
	http.HandleFunc("/api/create-license-request", handlers.CreateLicenseRequestHandler)

	// Пути к сертификату и ключу
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

	// Запуск HTTPS-сервера
	log.Println("HTTPS server started on port 8443...")
	err = http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
	}
}

package main

import (
	"LicenseApp/pkg/db"
	"LicenseApp/pkg/handlers"
	"LicenseApp/pkg/security"
	"log"
	"net/http"
)

func main() {
	// 1. Загрузка ключей
	err := security.LoadKeys("config/keys/private_key.pem", "config/keys/public_key.pem")
	if err != nil {
		log.Fatal("Error loading keys:", err)
	}

	// 2. Инициализация базы
	db.Init()
	db.Migrate()

	// 3. Маршруты
	http.HandleFunc("/license-requests", handlers.GetLicenseRequestsHandler)
	http.HandleFunc("/approve-license", handlers.ApproveLicenseRequestHandler)
	http.HandleFunc("/reject-license", handlers.RejectLicenseRequestHandler)

	log.Println("Server started on port 8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

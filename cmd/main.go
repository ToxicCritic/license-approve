package main

import (
	"LicenseApp/pkg/db"
	"LicenseApp/pkg/handlers"
	"log"
	"net/http"
)

func main() {
	db.Init()
	db.Migrate()

	licenseCheckHandler := handlers.LicenseCheckHandler{
		Checker: &db.LicenseDBChecker{DB: db.DB},
	}

	http.HandleFunc("/license-requests", handlers.GetLicenseRequestsHandler)
	http.HandleFunc("/approve-license", handlers.ApproveLicenseRequestHandler)
	http.HandleFunc("/reject-license", handlers.RejectLicenseRequestHandler)
	http.HandleFunc("/check-license", licenseCheckHandler.CheckLicenseHandler)

	// Сертификат и ключ (самоподписанные)
	certFile := "config/certs/server.crt"
	keyFile := "config/certs/server.key"

	log.Println("HTTPS server started on port 8443...")
	err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
	if err != nil {
		log.Fatal("Error starting HTTPS server:", err)
	}
}

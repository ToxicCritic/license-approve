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

	http.HandleFunc("/license-requests", handlers.GetLicenseRequestsHandler)
	http.HandleFunc("/approve-license", handlers.ApproveLicenseRequestHandler)
	http.HandleFunc("/reject-license", handlers.RejectLicenseRequestHandler)

	log.Println("Server started on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

package main

import (
	"LicenseApp/pkg/db"
	"LicenseApp/pkg/handlers"
	"LicenseApp/pkg/security"
	"fmt"
	"log"
	"net/http"
	"time"
)

const myUserID = 10

func main() {
	db.Init()
	db.Migrate()

	security.LoadKeys("config/keys/private_key.pem", "config/keys/public_key.pem")
	licenseChecker := &db.LicenseDBChecker{DB: db.DB}

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

	go func() {
		log.Println("HTTPS server started on port 8443...")
		err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
		if err != nil {
			log.Fatal("Error starting HTTPS server:", err)
		}
	}()

	time.Sleep(2 * time.Second)

	fmt.Println("=== First Program Start ===")

	// Проверяем, есть ли лицензия у myUserID
	hasLicense, err := licenseChecker.CheckLicense(myUserID)
	if err != nil {
		log.Fatalf("Failed to check license: %v", err)
	}
	if !hasLicense {
		fmt.Printf("No license found for user %d. Creating license request...\n", myUserID)
		// Здесь мы можем либо:
		// 1) вызвать db.CreateLicenseRequest(myUserID, "PUBLIC_KEY") напрямую,
		// 2) или отправить POST /license-requests (имитируя клиента).

		requestID, err := db.CreateLicenseRequest(myUserID, "MOCK_PUBLIC_KEY")
		if err != nil {
			log.Fatalf("Failed to create license request: %v", err)
		}
		fmt.Printf("License request #%d created, status='pending'.\n", requestID)

		fmt.Println()
		fmt.Println("Manager must now approve the request (ID:", requestID, ").")
		fmt.Println("Use Postman/cURL: POST http://localhost:8443/approve-license  { \"id\":", requestID, " }")
		fmt.Println("Waiting 60 seconds for approval...")

		time.Sleep(60 * time.Second)

		// Повторная проверка
		hasLicenseAgain, err := licenseChecker.CheckLicense(myUserID)
		if err != nil {
			log.Fatalf("Failed to check license (second time): %v", err)
		}
		if hasLicenseAgain {
			fmt.Println("License is now approved! Program can continue.")
		} else {
			fmt.Println("Still no license. Program remains in restricted mode or exits.")
		}
	} else {
		fmt.Printf("License for user %d found. Program can run fully.\n", myUserID)
	}

	fmt.Println("=== Program finished ===")
}

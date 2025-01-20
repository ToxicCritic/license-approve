// cmd/main.go
package main

import (
	"LicenseApp/pkg/db"
	"LicenseApp/pkg/handlers"
	"LicenseApp/pkg/security"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	myUserID         = 13
	checkInterval    = 10 * time.Second // Интервал между проверками лицензии
	maxCheckDuration = 5 * time.Minute  // Максимальное время ожидания одобрения лицензии
)

func main() {
	// Инициализация базы данных
	db.Init()
	db.Migrate()

	// Создание реализации LicenseChecker
	licenseChecker := &db.LicenseDBChecker{DB: db.DB}

	// Регистрация HTTP-обработчиков
	licenseCheckHandler := handlers.LicenseCheckHandler{
		Checker: licenseChecker,
	}

	// Регистрация обработчиков без аутентификации
	http.HandleFunc("/admin/license-requests", handlers.GetLicenseRequestsHandler)
	http.HandleFunc("/admin/approve-license", handlers.ApproveLicenseRequestHandler)
	http.HandleFunc("/admin/reject-license", handlers.RejectLicenseRequestHandler)
	http.HandleFunc("/check-license", licenseCheckHandler.CheckLicenseHandler)

	// Получение текущей рабочей директории
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}

	// Определение пути к директории конфигурации
	configPath := filepath.Join(cwd, "config")

	// Пути к сертификату и ключу
	certFile := filepath.Join(configPath, "certs", "server.crt")
	keyFile := filepath.Join(configPath, "certs", "server.key")
	log.Println("Certificate file path:", certFile)
	log.Println("Key file path:", keyFile)

	// Пути к ключам для подписания лицензий
	privateKeyFile := filepath.Join(configPath, "keys", "private_key.pem")
	publicKeyFile := filepath.Join(configPath, "keys", "public_key.pem")

	// Проверка существования файла сертификата
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificate not found at path: %s", certFile)
	} else if err != nil {
		log.Fatalf("Error checking certificate file: %v", err)
	}

	// Проверка существования файла ключа
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Key not found at path: %s", keyFile)
	} else if err != nil {
		log.Fatalf("Error checking key file: %v", err)
	}

	// Проверка существования приватного ключа
	if _, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
		log.Fatalf("Private key not found at path: %s", privateKeyFile)
	} else if err != nil {
		log.Fatalf("Error checking private key file: %v", err)
	}

	// Проверка существования публичного ключа
	if _, err := os.Stat(publicKeyFile); os.IsNotExist(err) {
		log.Fatalf("Public key not found at path: %s", publicKeyFile)
	} else if err != nil {
		log.Fatalf("Error checking public key file: %v", err)
	}

	// Загрузка ключей для подписания лицензий
	err = security.LoadKeys(privateKeyFile, publicKeyFile)
	if err != nil {
		log.Fatalf("Error loading security keys: %v", err)
	}

	// Запуск HTTPS-сервера в отдельной горутине
	go func() {
		log.Println("HTTPS server started on port 8443...")
		err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
		if err != nil {
			log.Fatal("Error starting HTTPS server:", err)
		}
	}()

	// Ожидание запуска сервера
	time.Sleep(2 * time.Second)

	fmt.Println("=== First Program Start ===")

	// Создание WaitGroup для ожидания завершения периодических проверок
	var wg sync.WaitGroup
	wg.Add(1)

	// Проверка наличия лицензии у пользователя
	hasLicense, err := licenseChecker.CheckLicense(myUserID)
	if err != nil {
		log.Fatalf("Failed to check license: %v", err)
	}
	if !hasLicense {
		// Проверка наличия уже существующей заявки на лицензию
		pendingRequest, err := db.GetPendingLicenseRequestByUserID(myUserID)
		if err != nil {
			log.Fatalf("Error checking existing license requests: %v", err)
		}

		if pendingRequest != nil {
			fmt.Printf("User %d already has a pending license request #%d.\n", myUserID, pendingRequest.ID)
			fmt.Println("Waiting for the request to be approved...")
		} else {
			fmt.Printf("License not found for user %d. Creating a license request...\n", myUserID)
			// Создание новой заявки на лицензию
			requestID, err := db.CreateLicenseRequest(myUserID, "MOCK_PUBLIC_KEY")
			if err != nil {
				log.Fatalf("Failed to create license request: %v", err)
			}
			fmt.Printf("License request #%d created with status='pending'.\n", requestID)

			fmt.Println()
			fmt.Println("Manager must approve the request (ID:", requestID, ").")
			fmt.Println("Use Postman/cURL: POST https://localhost:8443/approve-license  { \"id\":", requestID, " }")
			fmt.Println("Starting periodic license checks...")
		}

		// Запуск периодической проверки лицензии
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(checkInterval)
			defer ticker.Stop()

			timeout := time.After(maxCheckDuration)

			for {
				select {
				case <-ticker.C:
					hasLicenseNow, err := licenseChecker.CheckLicense(myUserID)
					if err != nil {
						log.Printf("Failed to check license: %v", err)
						continue
					}
					if hasLicenseNow {
						fmt.Println("License approved! The program can continue.")

						lic, err := db.GetLicenseByUserID(myUserID)
						if err != nil {
							log.Fatalf("Failed to get license by userID=%d: %v", myUserID, err)
						}

						isValid, err := security.VerifyLicenseSignature(lic.LicenseKey, lic.LicenseSignature)
						if err != nil {
							log.Fatalf("Error verifying signature: %v", err)
						}
						if !isValid {
							log.Println("License signature is INVALID. It might have been tampered with or incorrectly signed.")
						} else {
							log.Println("License signature is valid. User license confirmed!")
						}

						// Остановка периодических проверок
						return
					} else {
						log.Println("License still not approved. Continuing to check...")
					}
				case <-timeout:
					fmt.Println("License approval wait time has expired.")
					// Решить, что делать дальше: выйти из программы или оставить в ограниченном режиме
					os.Exit(1)
				}
			}
		}()
	} else {
		fmt.Printf("License found for user %d. The program can run fully.\n", myUserID)
	}

	fmt.Println("=== Program is running ===")

	// Ожидание завершения периодических проверок
	wg.Wait()

	fmt.Println("=== Program Finished ===")
}

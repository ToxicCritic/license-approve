package main

import (
	"example.com/licence-approval/client/pkg/config"
	"example.com/licence-approval/client/pkg/errors"
	"example.com/licence-approval/client/pkg/handlers"
	"example.com/licence-approval/client/pkg/utils"

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	checkInterval    = 10 * time.Second
	maxCheckDuration = 5 * time.Minute
)

func main() {
	fmt.Println("=== Client Started ===")

	// Получаем путь к бинарнику
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	// Формируем путь к config.json в той же папке
	configPath := filepath.Join(exeDir, "config.json")

	// Загружаем config.json
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load client config: %v", err)
	}

	// Проверяем LICENSE_SERVER_URL
	if cfg.LicenseServerURL == "" {
		log.Fatal("LICENSE_SERVER_URL is not set in config.json")
	}

	// Генерируем LicenseKey, если нет
	if cfg.LicenseKey == "" {
		licenseKey, err := utils.GenerateHexLicenseKey()
		if err != nil {
			log.Fatalf("Failed to generate license key: %v", err)
		}
		cfg.LicenseKey = licenseKey
		fmt.Printf("Generated License Key: %s\n", cfg.LicenseKey)

		// Сохраняем
		if err := config.SaveConfig(configPath, cfg); err != nil {
			log.Fatalf("Failed to save config with new license key: %v", err)
		}
	} else {
		fmt.Printf("Using existing License Key: %s\n", cfg.LicenseKey)
	}

	// Читаем сертификат сервера (например, в ../server/config/certs/server.crt)
	certPath := filepath.Join(exeDir, "../server/config/certs/server.crt")
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("Failed to read server certificate: %v", err)
	}

	// Создаём пул доверенных сертификатов
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append server certificate to CA pool")
	}

	// Настраиваем TLS
	tlsConfig := &tls.Config{RootCAs: caCertPool}

	// Создаём HTTP-клиент
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Запуск проверки лицензии
	go func() {
		defer wg.Done()

		hasLicense, message, err := handlers.CheckLicense(httpClient, cfg.LicenseServerURL, cfg.LicenseKey)
		if err != nil {
			log.Printf("Failed to check license: %v", err)
			return
		}

		if hasLicense {
			fmt.Println("License is active. The client can proceed.")
			return
		} else {
			switch message {
			case "License is active.":
				fmt.Println("License is active. The client can proceed.")
				return
			case "License request is pending.":
				log.Println("License request is pending. Waiting for approval...")
			case "License request has been rejected.":
				log.Println("Your license request has been rejected by the administrator. Please contact support.")
				return
			default:
				log.Println("License is not active. Creating a new license request...")
			}
		}

		// Создаём заявку
		requestID, err := handlers.CreateLicenseRequest(httpClient, cfg.LicenseServerURL, cfg.LicenseKey)
		if err != nil {
			if reqErr, ok := err.(*errors.LicenseRequestExistsError); ok {
				log.Printf("License request already exists with ID %d. Waiting for approval...", reqErr.RequestID)
			} else {
				log.Printf("Failed to create license request: %v", err)
				return
			}
		} else {
			log.Printf("License request #%d created. Waiting for approval...", requestID)
		}

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		timeoutChan := time.After(maxCheckDuration)

		for {
			select {
			case <-ticker.C:
				hasLicenseNow, msg, err := handlers.CheckLicense(httpClient, cfg.LicenseServerURL, cfg.LicenseKey)
				if err != nil {
					if _, ok := err.(*errors.LicenseRejectedError); ok {
						fmt.Println("Your license request has been rejected by the administrator.")
						return
					}
					log.Printf("Failed to check license: %v", err)
					continue
				}
				if hasLicenseNow {
					fmt.Println("License approved! The client can proceed.")
					return
				} else {
					log.Printf("License status: %s. Continuing to check...", msg)
				}

			case <-timeoutChan:
				fmt.Println("The waiting time for license approval has expired.")
				os.Exit(1)
			}
		}
	}()

	fmt.Println("=== Client is running ===")

	// Ждём завершения проверки
	wg.Wait()

	fmt.Println("=== Client Finished ===")
}

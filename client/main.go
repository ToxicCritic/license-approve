// client/main.go

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/licence-approval/client/pkg/config"
	"example.com/licence-approval/client/pkg/handlers"
	"example.com/licence-approval/client/pkg/utils"

	"example.com/licence-approval/client/pkg/errors"
)

const (
	checkInterval    = 10 * time.Second // Интервал между проверками лицензии
	maxCheckDuration = 5 * time.Minute  // Максимальное время ожидания одобрения лицензии
)

func main() {
	fmt.Println("=== Client Started ===")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Если LicenseKey пустой, генерируем новый и сохраняем его
	if cfg.LicenseKey == "" {
		licenseKey, err := utils.GenerateHexLicenseKey()
		if err != nil {
			log.Fatalf("Failed to generate license key: %v", err)
		}
		cfg.LicenseKey = licenseKey
		fmt.Printf("Generated License Key: %s\n", cfg.LicenseKey)

		// Сохраняем новый LicenseKey в конфигурационный файл
		if err := config.SaveConfig("config.json", cfg); err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}
	} else {
		fmt.Printf("Using existing License Key: %s\n", cfg.LicenseKey)
	}

	// Путь к сертификату сервера
	certPath := filepath.Join(getCWD(), "../server/config/certs/server.crt")

	// Загрузка серверного сертификата
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("Failed to read server certificate: %v", err)
	}

	// Создание пула доверенных сертификатов
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append server certificate to CA pool")
	}

	// Настройка TLS-конфигурации
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	// Создание пользовательского транспортного слоя с TLS-конфигурацией
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Создание HTTP-клиента с пользовательским транспортом
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Создание WaitGroup для ожидания завершения периодических проверок
	var wg sync.WaitGroup
	wg.Add(1)

	// Запуск периодической проверки лицензии
	go func() {
		defer wg.Done()

		// Проверяем статус лицензии
		hasLicense, message, err := handlers.CheckLicense(client, cfg.LicenseServerURL, cfg.LicenseKey)
		if err != nil {
			// Обработка ошибки, например, вывод сообщения и выход из приложения
			log.Printf("Failed to check license: %v", err)
			return
		}

		if hasLicense {
			fmt.Println("License is active. The client can proceed.")
			// Здесь можно добавить дальнейшую логику работы клиента
			return
		} else {
			// Обработка различных сообщений
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

		// Создаем заявку на лицензию
		requestID, err := handlers.CreateLicenseRequest(client, cfg.LicenseServerURL, cfg.LicenseKey)
		if err != nil {
			// Проверяем, была ли ошибка из-за существующей заявки
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

		timeout := time.After(maxCheckDuration)

		for {
			select {
			case <-ticker.C:
				// Проверяем статус лицензии
				hasLicenseNow, message, err := handlers.CheckLicense(client, cfg.LicenseServerURL, cfg.LicenseKey)
				if err != nil {
					// Проверяем, является ли ошибка LicenseRejectedError
					if _, ok := err.(*errors.LicenseRejectedError); ok {
						fmt.Println("Your license request has been rejected by the administrator. Please contact support.")
						return
					}
					log.Printf("Failed to check license: %v", err)
					continue
				}
				if hasLicenseNow {
					fmt.Println("License approved! The client can proceed.")
					// Здесь можно добавить дальнейшую логику работы клиента
					return
				} else {
					log.Printf("License status: %s. Continuing to check...", message)
				}
			case <-timeout:
				fmt.Println("The waiting time for license approval has expired.")
				// Решите, что делать дальше: выйти из программы или оставить в ограниченном режиме
				os.Exit(1)
			}
		}
	}()

	fmt.Println("=== Client is running ===")

	// Ожидание завершения периодических проверок
	wg.Wait()

	fmt.Println("=== Client Finished ===")
}

// loadConfig загружает конфигурацию из указанного JSON файла
func loadConfig(path string) (*config.Config, error) {
	return config.LoadConfig(path)
}

// saveConfig сохраняет конфигурацию в указанный JSON файл
func saveConfig(path string, cfg *config.Config) error {
	return config.SaveConfig(path, cfg)
}

// getCWD возвращает текущую рабочую директорию
func getCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	return cwd
}

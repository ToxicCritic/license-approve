// client/main.go
package main

import (
	"LicenseApp/client/pkg/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"crypto/tls"
	"crypto/x509"

	"github.com/joho/godotenv"
)

const (
	checkInterval    = 10 * time.Second // Интервал между проверками лицензии
	maxCheckDuration = 5 * time.Minute  // Максимальное время ожидания одобрения лицензии
)

func main() {
	fmt.Println("=== Client Started ===")

	// Загрузка переменных окружения из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using default configuration.")
	}

	// Получение URL сервера из переменных окружения
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "https://localhost:8443"
	}

	// Получение ID пользователя из переменных окружения
	userIDStr := os.Getenv("USER_ID")
	if userIDStr == "" {
		userIDStr = "13"
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Fatalf("Invalid USER_ID: %v", err)
	}

	// Путь к сертификату сервера
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	// Предполагается, что server.crt находится в ../server/config/certs/
	certPath := filepath.Join(cwd, "../server/config/certs/server.crt")

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
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		timeout := time.After(maxCheckDuration)

		for {
			select {
			case <-ticker.C:
				// Вызов API для проверки лицензии
				hasLicense, err := handlers.CheckLicense(client, serverURL, userID)
				if err != nil {
					log.Printf("Failed to check license: %v", err)
					continue
				}
				if hasLicense {
					fmt.Println("License approved! The client can continue.")
					// Здесь можно добавить дальнейшую логику работы клиента
					return
				} else {
					log.Println("License not approved yet. Continuing to check...")
				}
			case <-timeout:
				fmt.Println("License approval wait time has expired.")
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

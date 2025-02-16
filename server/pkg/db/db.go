package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	// Считываем переменные окружения для подключения к БД
	dbUser := getEnvOrDefault("DB_USER", "license_user")
	dbPass := getEnvOrDefault("DB_PASS", "yourpassword")
	dbName := getEnvOrDefault("DB_NAME", "license_db")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")

	// Формируем строку подключения
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPass, dbName, dbHost, dbPort, dbSSLMode,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
}

// Migrate выполняет миграции для создания таблиц
func Migrate() {
	createLicensesTable := `
	CREATE TABLE IF NOT EXISTS licenses (
		id SERIAL PRIMARY KEY,
		license_key VARCHAR(100) NOT NULL UNIQUE,
		license_signature TEXT,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		tag INT NOT NULL
	);
	`

	createRequestsTable := `
	CREATE TABLE IF NOT EXISTS license_requests (
		id SERIAL PRIMARY KEY,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		license_key VARCHAR(100) NOT NULL UNIQUE
	);
	`

	queries := []string{
		createLicensesTable,
		createRequestsTable,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing migration: %v", err)
		}
	}

	log.Println("Database tables created successfully!")
}

// getEnvOrDefault пытается получить значение переменной окружения или возвращает значение по умолчанию
func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

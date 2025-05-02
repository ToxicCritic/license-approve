package db

import (
	"database/sql"
	"fmt"
	"log"

	"server/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) {
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode,
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

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	queries := []string{
		createLicensesTable,
		createRequestsTable,
		createUsersTable,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing migration: %v", err)
		}
	}

	log.Println("Database tables created successfully!")
}

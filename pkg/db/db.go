package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	var err error
	connStr := "user=license_user password=yourpassword dbname=license_db sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database is not reachable: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
}

// Миграции для создания таблиц
func Migrate() {
	// Создание таблицы пользователей
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		public_key TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Создание таблицы лицензий
	createLicensesTable := `
	CREATE TABLE IF NOT EXISTS licenses (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    license_key TEXT NOT NULL UNIQUE,
    license_signature TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	// Создание таблицы запросов на лицензии
	createRequestsTable := `
	CREATE TABLE IF NOT EXISTS license_requests (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		public_key TEXT NOT NULL,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	queries := []string{createUsersTable, createLicensesTable, createRequestsTable}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Error executing migration: %v", err)
		}
	}

	fmt.Println("Database tables created successfully!")
}

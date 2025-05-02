package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config хранит настройки mock OAuth2.0 сервера и параметры подключения к БД
// Все значения читаются из переменных окружения с возможностью указания дефолтных путей
// для сертификатов, расположенных рядом с бинарником.
type Config struct {
	Addr     string // Адрес запуска (переменная OAUTH_ADDR, default ":8081")
	CertFile string // Путь к сертификату (переменная OAUTH_CERT_FILE или default)
	KeyFile  string // Путь к приватному ключу (переменная OAUTH_KEY_FILE или default)

	DBUser    string // DB_USER
	DBPass    string // DB_PASS
	DBName    string // DB_NAME
	DBHost    string // DB_HOST
	DBPort    string // DB_PORT
	DBSSLMode string // DB_SSLMODE
}

// getEnvOrDefault возвращает значение переменной окружения или defaultVal, если она не задана
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// LoadConfig загружает конфигурацию из переменных окружения и определяет пути к файлам сертификатов
func LoadConfig() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	cfg := &Config{
		Addr:     getEnvOrDefault("OAUTH_ADDR", ":8081"),
		CertFile: getEnvOrDefault("OAUTH_CERT_FILE", filepath.Join(exeDir, "certs", "mock-oauth.crt")),
		KeyFile:  getEnvOrDefault("OAUTH_KEY_FILE", filepath.Join(exeDir, "certs", "mock-oauth.key")),

		DBUser:    getEnvOrDefault("DB_USER", "license_user"),
		DBPass:    getEnvOrDefault("DB_PASS", "yourpassword"),
		DBName:    getEnvOrDefault("DB_NAME", "license_db"),
		DBHost:    getEnvOrDefault("DB_HOST", "db"),
		DBPort:    getEnvOrDefault("DB_PORT", "5432"),
		DBSSLMode: getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	return cfg, nil
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config представляет структуру конфигурации клиента
type Config struct {
	LicenseServerURL string `json:"LICENSE_SERVER_URL"`
	LicenseKey       string `json:"LICENSE_KEY"`
	// Добавьте другие необходимые поля, если есть
}

// LoadConfig загружает конфигурацию из указанного JSON файла
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Проверка обязательных переменных
	if cfg.LicenseServerURL == "" {
		return nil, fmt.Errorf("LICENSE_SERVER_URL is not set")
	}

	return &cfg, nil
}

// SaveConfig сохраняет конфигурацию в указанный JSON файл
func SaveConfig(path string, cfg *Config) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config to JSON: %w", err)
	}

	return nil
}

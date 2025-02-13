package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Addr     string // Адрес запуска (:8081)
	CertFile string // Путь к mock-oauth.crt
	KeyFile  string // Путь к mock-oauth.key
}

// LoadConfig пытается найти сертификаты рядом с бинарником,
// либо можно дополнить чтение из ENV, yaml и т.д.
func LoadConfig() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	cfg := &Config{
		Addr:     ":8081",
		CertFile: filepath.Join(exeDir, "certs", "mock-oauth.crt"),
		KeyFile:  filepath.Join(exeDir, "certs", "mock-oauth.key"),
	}
	return cfg, nil
}

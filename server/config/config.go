package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	OAuthClientID     string `mapstructure:"OAUTH_CLIENT_ID"`
	OAuthClientSecret string `mapstructure:"OAUTH_CLIENT_SECRET"`
	OAuthRedirectURL  string `mapstructure:"OAUTH_REDIRECT_URL"`
	OAuthAuthURL      string `mapstructure:"OAUTH_AUTH_URL"`
	OAuthTokenURL     string `mapstructure:"OAUTH_TOKEN_URL"`
	SessionSecret     string `mapstructure:"SESSION_SECRET"`
	PrivateKeyPath    string `mapstructure:"PRIVATE_KEY_PATH"`
	PublicKeyPath     string `mapstructure:"PUBLIC_KEY_PATH"`
	CertFile          string `mapstructure:"CERT_FILE"`
	KeyFile           string `mapstructure:"KEY_FILE"`

	DBUser    string `mapstructure:"DB_USER"`
	DBPass    string `mapstructure:"DB_PASS"`
	DBName    string `mapstructure:"DB_NAME"`
	DBHost    string `mapstructure:"DB_HOST"`
	DBPort    string `mapstructure:"DB_PORT"`
	DBSSLMode string `mapstructure:"DB_SSLMODE"`
}

// LoadConfig загружает конфигурацию из файла .env, расположенного рядом с бинарником.
func LoadConfig() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env in %s, using environment: %v\n", exeDir, err)
	}
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Проверка обязательных переменных для OAuth/TLS
	if cfg.OAuthClientID == "" || cfg.OAuthClientSecret == "" || cfg.OAuthRedirectURL == "" ||
		cfg.OAuthAuthURL == "" || cfg.OAuthTokenURL == "" || cfg.SessionSecret == "" ||
		cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" ||
		cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("missing required OAuth/TLS config fields")
	}

	// Проверка обязательных переменных для базы данных
	if cfg.DBUser == "" || cfg.DBPass == "" || cfg.DBName == "" || cfg.DBHost == "" ||
		cfg.DBPort == "" || cfg.DBSSLMode == "" {
		return nil, fmt.Errorf("missing required DB config fields")
	}

	return &cfg, nil
}

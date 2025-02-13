package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	OAuthClientID     string `mapstructure:"OAUTH_CLIENT_ID"`
	OAuthClientSecret string `mapstructure:"OAUTH_CLIENT_SECRET"`
	OAuthRedirectURL  string `mapstructure:"OAUTH_REDIRECT_URL"`
	OAuthAuthURL      string `mapstructure:"OAUTH_AUTH_URL"`
	OAuthTokenURL     string `mapstructure:"OAUTH_TOKEN_URL"`

	SessionSecret string `mapstructure:"SESSION_SECRET"`

	PrivateKeyPath string `mapstructure:"PRIVATE_KEY_PATH"`
	PublicKeyPath  string `mapstructure:"PUBLIC_KEY_PATH"`

	CertFile string `mapstructure:"CERT_FILE"`
	KeyFile  string `mapstructure:"KEY_FILE"`
}

// LoadConfig загружает конфигурацию с использованием viper из .env файла
func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env") // Указываем файл конфигурации
	viper.SetConfigType("env")  // Тип файла конфигурации

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
	}

	// Автоматическое считывание переменных окружения
	viper.AutomaticEnv()

	// Считывание конфигурации в структуру
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("не удалось декодировать конфигурацию в структуру: %w", err)
	}

	// Проверка обязательных переменных
	if cfg.OAuthClientID == "" || cfg.OAuthClientSecret == "" || cfg.OAuthRedirectURL == "" ||
		cfg.OAuthAuthURL == "" || cfg.OAuthTokenURL == "" || cfg.SessionSecret == "" ||
		cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" ||
		cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("отсутствуют обязательные параметры конфигурации")
	}

	return &cfg, nil
}

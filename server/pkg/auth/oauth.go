package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"server/config"

	"golang.org/x/oauth2"
)

var OAuthConfig *oauth2.Config

// InitOAuthConfig инициализирует OAuth2.0 конфигурацию используя переданную структуру Config
func InitOAuthConfig(cfg *config.Config) {
	OAuthConfig = &oauth2.Config{
		ClientID:     cfg.OAuthClientID,
		ClientSecret: cfg.OAuthClientSecret,
		RedirectURL:  cfg.OAuthRedirectURL,
		Scopes:       []string{"read", "write"}, // какие нужно??
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.OAuthAuthURL,
			TokenURL: cfg.OAuthTokenURL,
		},
	}
}

// ExchangeCodeForToken обменивает авторизационный код на токен доступа,
// игнорируя проверку сертификата (InsecureSkipVerify) для локальной отладки.
func ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureClient := &http.Client{Transport: insecureTransport}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, insecureClient)

	token, err := OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}
	return token, nil
}

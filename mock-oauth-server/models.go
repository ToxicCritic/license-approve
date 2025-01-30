// mock-oauth-server/models.go

package main

import "time"

// User представляет пользователя в системе
type User struct {
	ID       string
	Username string
	Password string // В реальном приложении пароли хранятся в зашифрованном виде
}

// AuthorizationCode представляет авторизационный код
type AuthorizationCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Expiry      time.Time
}

// AccessToken представляет токен доступа
type AccessToken struct {
	Token     string
	UserID    string
	ClientID  string
	Expiry    time.Time
	TokenType string
}

// RefreshToken представляет токен обновления
type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

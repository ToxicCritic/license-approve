package main

import "time"

// Представляет пользователя в системе
type User struct {
	ID       string
	Username string
	Password string
}

// Представляет авторизационный код
type AuthorizationCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Expiry      time.Time
}

// Представляет токен доступа
type AccessToken struct {
	Token     string
	UserID    string
	ClientID  string
	Expiry    time.Time
	TokenType string
}

// Представляет токен обновления
type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

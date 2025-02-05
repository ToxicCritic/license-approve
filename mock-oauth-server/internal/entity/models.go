// internal/entity/models.go
package entity

import "time"

type User struct {
	ID       string
	Username string
	Password string
}

type Client struct {
	ID     string
	Secret string
	Domain string
	Public bool
	UserID string
}

type AccessToken struct {
	Token     string
	UserID    string
	ClientID  string
	Expiry    time.Time
	TokenType string
	Scope     string
}

type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

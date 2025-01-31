// mock-oauth-server/internal/models/models.go

package models

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// User represents a user (simple version)
type User struct {
	Username string
	Password string
}

// AuthorizationCode for the code flow
type AuthorizationCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
	Expiry      time.Time
}

// AccessToken represents an issued access token
type AccessToken struct {
	Token        string
	UserID       string
	ClientID     string
	TokenType    string
	Expiry       time.Time
	RefreshToken string
}

// RefreshToken representation
type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

// Global in-memory stores
var (
	Users         = map[string]*User{}              // userStore
	AuthCodes     = map[string]*AuthorizationCode{} // authCodes
	AccessTokens  = map[string]*AccessToken{}       // accessTokens
	RefreshTokens = map[string]*RefreshToken{}      // refreshTokens
	InMemoryMutex sync.Mutex
)

// InitStore initializes with one default user, etc.
func InitStore() {
	Users["demo"] = &User{
		Username: "demo",
		Password: "demo123",
	}
}

// GenerateRandomString returns a random base64url string of length n
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

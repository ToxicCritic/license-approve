package inmem

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Entities
type User struct {
	ID       string
	Username string
	Password string
}

type Group struct {
	Name string
}

type AuthorizationCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Expiry      time.Time
}

type AccessToken struct {
	Token     string
	UserID    string
	ClientID  string
	Expiry    time.Time
	TokenType string
}

type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

// Store

type Store struct {
	Users              map[string]*User
	Groups             map[string]*Group
	AuthorizationCodes map[string]*AuthorizationCode
	AccessTokens       map[string]*AccessToken
	RefreshTokens      map[string]*RefreshToken
}

func NewStore() *Store {
	s := &Store{
		Users:              make(map[string]*User),
		Groups:             make(map[string]*Group),
		AuthorizationCodes: make(map[string]*AuthorizationCode),
		AccessTokens:       make(map[string]*AccessToken),
		RefreshTokens:      make(map[string]*RefreshToken),
	}
	// дефолт user-1
	s.Users["user-1"] = &User{
		ID:       "user-1",
		Username: "admin",
		Password: "password",
	}
	// пример группы
	s.Groups["SuRtAdmin"] = &Group{Name: "SuRtAdmin"}
	return s
}

func (s *Store) GenerateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

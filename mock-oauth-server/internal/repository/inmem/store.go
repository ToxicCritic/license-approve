package inmem

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Entities
type User struct {
	Username     string
	PasswordHash []byte
	CreatedAt    time.Time
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

func NewStore(db *sql.DB) *Store {
	s := &Store{
		Users:              make(map[string]*User),
		AuthorizationCodes: make(map[string]*AuthorizationCode),
		AccessTokens:       make(map[string]*AccessToken),
		RefreshTokens:      make(map[string]*RefreshToken),
	}

	rows, err := db.Query("SELECT login, password_hash, created_at FROM users")
	if err != nil {
		log.Fatalf("Failed to query users table: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var login, hash string
		var createdAt time.Time
		if err := rows.Scan(&login, &hash, &createdAt); err != nil {
			log.Fatalf("Error scanning user row: %v", err)
		}

		if _, err := bcrypt.Cost([]byte(hash)); err != nil {
			log.Printf("Warning: invalid hash for user %s: %v", login, err)
			continue
		}
		s.Users[login] = &User{
			Username:     login,
			PasswordHash: []byte(hash),
			CreatedAt:    createdAt,
		}
	}

	if len(s.Users) == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		s.Users["admin"] = &User{Username: "admin", PasswordHash: hash, CreatedAt: time.Now()}
	}

	return s
}

func (s *Store) GenerateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

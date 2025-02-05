// internal/service/jwtaccess/jwt.go
package jwtaccess

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mock-authserver/internal/entity"
	"time"

	"github.com/golang-jwt/jwt"
)

type Store interface {
	SaveAccessToken(at *entity.AccessToken)
	GetAccessToken(token string) (*entity.AccessToken, bool)
	SaveRefreshToken(rt *entity.RefreshToken)
	GetRefreshToken(token string) (*entity.RefreshToken, bool)
	DeleteRefreshToken(token string)
	FindClient(clientID, secret string) (*entity.Client, bool)
}

type JWTService struct {
	store     Store
	jwtSecret []byte
}

func NewJWTService(store Store, secret string) *JWTService {
	return &JWTService{
		store:     store,
		jwtSecret: []byte(secret),
	}
}

func (js *JWTService) GenerateTokenPair(userID, clientID string) (string, string, error) {
	now := time.Now()
	// Создаём claims
	claims := jwt.MapClaims{
		"sub": userID,
		"aud": clientID,
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(js.jwtSecret)
	if err != nil {
		return "", "", err
	}
	js.store.SaveAccessToken(&entity.AccessToken{
		Token:     signed,
		UserID:    userID,
		ClientID:  clientID,
		Expiry:    now.Add(time.Hour),
		TokenType: "Bearer",
	})

	// refresh token
	r, err := randomString(32)
	if err != nil {
		return "", "", err
	}
	js.store.SaveRefreshToken(&entity.RefreshToken{
		Token:    r,
		UserID:   userID,
		ClientID: clientID,
		Expiry:   now.Add(24 * time.Hour),
	})
	return signed, r, nil
}

// RefreshToken ...
func (js *JWTService) RefreshToken(rToken string) (string, string, error) {
	rt, ok := js.store.GetRefreshToken(rToken)
	if !ok {
		return "", "", fmt.Errorf("invalid refresh token")
	}
	if time.Now().After(rt.Expiry) {
		return "", "", fmt.Errorf("refresh token expired")
	}
	// Выдаём новую пару
	at, rt2, err := js.GenerateTokenPair(rt.UserID, rt.ClientID)
	if err != nil {
		return "", "", err
	}
	js.store.DeleteRefreshToken(rToken)
	return at, rt2, nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Дополнительно методы для валидации JWT из access_token, если надо

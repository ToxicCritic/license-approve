// mock-oauth-server/utils.go

package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// generateRandomString генерирует случайную строку заданной длины
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// validateClient проверяет, существует ли клиент с данным ID и секретом
func validateClient(clientID, clientSecret string) bool {
	// В реальном приложении клиентские данные хранятся в базе данных
	// Здесь для примера используем статический клиент
	return clientID == "mock-client-id" && clientSecret == "mock-client-secret"
}

// findUser находит пользователя по имени пользователя и паролю
func findUser(username, password string) (*User, error) {
	// В реальном приложении пользователи хранятся в базе данных
	// Здесь для примера используем статического пользователя
	if username == "admin" && password == "password" {
		return &User{
			ID:       "user-1",
			Username: "admin",
			Password: "password",
		}, nil
	}
	return nil, errors.New("invalid credentials")
}

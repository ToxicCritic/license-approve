package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// Генерирует случайную строку заданной длины
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func validateClient(clientID, clientSecret string) (bool, bool) {
	// В реальном приложении логика должна включать запрос к базе данных
	// Здесь для примера используем статический "000000" / "999999".
	validID := (clientID == "000000")
	validSecret := (clientSecret == "999999")
	return validID, validSecret
}

// Находит пользователя по имени пользователя и паролю
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

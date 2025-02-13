package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

// Генерирует лицензионный ключ (SHA-256) на основе библиотеки denisbrodbeck/machineid
// и возвращает его в шестнадцатеричном виде.
func GenerateHexLicenseKey() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %v", err)
	}

	hashSum := sha256.Sum256([]byte(id))

	return hex.EncodeToString(hashSum[:]), nil
}

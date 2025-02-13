package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// Приватный ключ (RSA)
var privateKey *rsa.PrivateKey

// Публичный ключ (RSA)
var publicKey *rsa.PublicKey

// Читает приватный и публичный ключи из PEM-файлов (PKCS#8 + PKIX)
func LoadKeys(privateKeyPath, publicKeyPath string) error {
	// ----- ЗАГРУЗКА ПРИВАТНОГО КЛЮЧА -----
	log.Printf("Attempting to open private key at: %s", privateKeyPath)
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %v", err)
	}

	privBlock, _ := pem.Decode(privBytes)
	if privBlock == nil {
		return fmt.Errorf("no PEM block found in private key file")
	}

	// Парсим PKCS#8 приватный ключ
	parsedKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse PKCS#8 private key: %v", err)
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("parsed private key is not an RSA key")
	}
	privateKey = rsaKey

	// ----- ЗАГРУЗКА ПУБЛИЧНОГО КЛЮЧА -----
	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read public key file: %v", err)
	}

	pubBlock, _ := pem.Decode(pubBytes)
	if pubBlock == nil {
		return fmt.Errorf("no PEM block found in public key file")
	}

	// Парсим PKIX приватный ключ
	parsedPubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse public key: %v", err)
	}
	switch pub := parsedPubKey.(type) {
	case *rsa.PublicKey:
		publicKey = pub
	default:
		return fmt.Errorf("public key is not an RSA key")
	}

	log.Println("Keys loaded successfully!")
	return nil
}

// -----------------------------------------------------------------------------
// ЦИФРОВАЯ ПОДПИСЬ
// -----------------------------------------------------------------------------

// Cоздает цифровую подпись (RSA) для строкового licenseKey
func SignLicense(licenseKey string) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("private key not loaded")
	}
	hash := sha256.New()
	hash.Write([]byte(licenseKey))
	digest := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, 0, digest)
	if err != nil {
		return "", fmt.Errorf("failed to sign license: %v", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// Проверяет подлинность licenseKey, используя полученную подпись и публичный ключ
func VerifyLicenseSignature(licenseKey, signatureB64 string) (bool, error) {
	if publicKey == nil {
		return false, fmt.Errorf("public key not loaded")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, fmt.Errorf("invalid base64 signature: %v", err)
	}

	hash := sha256.New()
	hash.Write([]byte(licenseKey))
	digest := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(publicKey, 0, digest, signature)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// -----------------------------------------------------------------------------
// КЛИЕНТСКОЕ ШИФРОВАНИЕ (Публичным ключом) и Расшифровка (Приватным ключом)
// -----------------------------------------------------------------------------

// Шифрует исходные данные (plainData) публичным ключом (RSA OAEP).
// Возвращает Base64-строку для удобства передачи.
func EncryptDataWithPublicKey(plainData []byte) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key not loaded")
	}

	cipherData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plainData, nil)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}

	return base64.StdEncoding.EncodeToString(cipherData), nil
}

// Расшифровывает Base64-строку, используя приватный ключ сервера
// (RSA OAEP). Возвращает исходные байты plainData.
func DecryptDataWithPrivateKey(base64Cipher string) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("server private key not loaded")
	}

	cipherData, err := base64.StdEncoding.DecodeString(base64Cipher)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 cipher data: %v", err)
	}

	plainData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, cipherData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}

	return plainData, nil
}

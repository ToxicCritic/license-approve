package models

import "time"

// Представляет структуру лицензии
type License struct {
	ID               int       `json:"id"`
	LicenseKey       string    `json:"license_key"`
	LicenseSignature string    `json:"license_signature"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	Tag              int       `json:"tag"`
	ApprovedBy       string    `json:"approved_by"`
}

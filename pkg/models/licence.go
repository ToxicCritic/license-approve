package models

import "time"

// License представляет саму лицензию
type License struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	LicenseKey       string    `json:"license_key"`
	LicenseSignature string    `json:"license_signature"`
	Status           string    `json:"status"` // active, expired, approvedJ
	IssuedAt         time.Time `json:"issued_at"`
}

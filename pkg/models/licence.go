package models

import "time"

// License представляет саму лицензию
type License struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	LicenseKey string    `json:"license_key"`
	Status     string    `json:"status"` // Статус лицензии (active, expired)
	IssuedAt   time.Time `json:"issued_at"`
}

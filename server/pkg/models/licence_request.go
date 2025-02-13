package models

import "time"

// Представляет структуру запроса на лицензию
type LicenseRequest struct {
	ID         int       `json:"id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	LicenseKey string    `json:"license_key"`
}

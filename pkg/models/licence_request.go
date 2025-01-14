package models

import "time"

// LicenseRequest представляет запрос на лицензию
type LicenseRequest struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PublicKey string    `json:"public_key"`
	Status    string    `json:"status"` // Статус запроса (pending, approved, rejected)
	CreatedAt time.Time `json:"created_at"`
}

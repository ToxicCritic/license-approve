package db

import (
	"LicenseApp/pkg/licensegen"
	"LicenseApp/pkg/models"
	"LicenseApp/pkg/security"
	"database/sql"
	"fmt"
	"log"
)

// Создает заявку на лицензию
func CreateLicenseRequest(userID int, publicKey string) (int, error) {
	hasPending, err := HasPendingLicenseRequest(userID)
	if err != nil {
		return 0, err
	}
	if hasPending {
		return 0, fmt.Errorf("pending license request already exists for user %d", userID)
	}

	var newID int
	err = DB.QueryRow(`
        INSERT INTO license_requests (user_id, public_key, status)
        VALUES ($1, $2, 'pending')
        RETURNING id
    `, userID, publicKey).Scan(&newID)

	if err != nil {
		return 0, err
	}
	return newID, nil
}

// Получает заявки на рассмотрении (status = 'pending')
func GetLicenseRequests() ([]models.LicenseRequest, error) {
	rows, err := DB.Query("SELECT * FROM license_requests WHERE status = 'pending' ORDER BY created_at DESC")
	if err != nil {
		log.Println("Error fetching license requests:", err)
		return nil, err
	}
	defer rows.Close()

	var requests []models.LicenseRequest
	for rows.Next() {
		var req models.LicenseRequest
		if err := rows.Scan(&req.ID, &req.UserID, &req.PublicKey, &req.Status, &req.CreatedAt); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// Одобряет заявку
func ApproveLicenseRequest(id int) error {
	var userID int
	var publicKey string

	err := DB.QueryRow("SELECT user_id, public_key FROM license_requests WHERE id = $1", id).
		Scan(&userID, &publicKey)
	if err != nil {
		return err
	}

	licenseKey := licensegen.GenerateHexLicenseKey(publicKey)

	signature, err := security.SignLicense(licenseKey)
	if err != nil {
		return fmt.Errorf("failed to sign license: %v", err)
	}

	_, err = DB.Exec(`INSERT INTO licenses (user_id, license_key, license_signature, status) 
					  VALUES ($1, $2, $3, 'approved')`,
		userID, licenseKey, signature)
	if err != nil {
		return err
	}

	_, err = DB.Exec("UPDATE license_requests SET status = 'approved' WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

// Отклоняет заявку
func RejectLicenseRequest(id int) error {
	_, err := DB.Exec("UPDATE license_requests SET status = 'rejected' WHERE id = $1", id)
	if err != nil {
		log.Println("Error rejecting license request:", err)
		return err
	}
	return nil
}

// Получает лицензию пользователя по его ID
func GetLicenseByUserID(userID int) (*models.License, error) {
	row := DB.QueryRow(`
        SELECT id, user_id, license_key, license_signature, status 
        FROM licenses
        WHERE user_id = $1 AND status = 'approved'
        LIMIT 1
    `, userID)

	var lic models.License
	err := row.Scan(&lic.ID, &lic.UserID, &lic.LicenseKey, &lic.LicenseSignature, &lic.Status)
	if err != nil {
		log.Println("Error fetching license:", err)
		return nil, err
	}

	return &lic, nil
}

// Проверяет, есть ли у пользователя уже заявка со статусом 'pending'
func HasPendingLicenseRequest(userID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM license_requests
			WHERE user_id = $1 AND status = 'pending'
			LIMIT 1
		)
	`
	err := DB.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking pending license request for user %d: %v", userID, err)
		return false, err
	}
	return exists, nil
}

// Возвращает pending-запрос для пользователя, если он существует
func GetPendingLicenseRequestByUserID(userID int) (*models.LicenseRequest, error) {
	var req models.LicenseRequest
	query := `
		SELECT id, user_id, public_key, status, created_at
		FROM license_requests
		WHERE user_id = $1 AND status = 'pending'
		LIMIT 1
	`
	err := DB.QueryRow(query, userID).Scan(&req.ID, &req.UserID, &req.PublicKey, &req.Status, &req.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Заявка не найдена
		}
		log.Printf("Error fetching pending license request for user %d: %v", userID, err)
		return nil, err
	}
	return &req, nil
}

func GetLicenseRequestByID(requestID int) (*models.LicenseRequest, error) {
	var req models.LicenseRequest
	query := `
		SELECT id, user_id, public_key, status, created_at
		FROM license_requests
		WHERE id = $1
		LIMIT 1
	`
	err := DB.QueryRow(query, requestID).Scan(&req.ID, &req.UserID, &req.PublicKey, &req.Status, &req.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Заявка не найдена
		}
		log.Printf("Error fetching license request %d: %v", requestID, err)
		return nil, err
	}
	return &req, nil
}

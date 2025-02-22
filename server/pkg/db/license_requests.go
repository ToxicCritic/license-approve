package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"server/pkg/models"
	"server/pkg/security"

	"github.com/lib/pq"
)

func CreateLicenseRequest(licenseKey string) (*models.LicenseRequest, error) {
	query := `
        INSERT INTO license_requests (license_key, status)
        VALUES ($1, 'pending')
        RETURNING id, status, created_at, license_key;
    `

	var request models.LicenseRequest
	err := DB.QueryRow(query, licenseKey).Scan(
		&request.ID,
		&request.Status,
		&request.CreatedAt,
		&request.LicenseKey,
	)
	if err != nil {
		// Проверка на нарушение уникальности license_key
		if isUniqueViolation(err) {
			log.Printf("License key %s already exists. Attempting to fetch existing request.", licenseKey)
			// Попытка получить существующую заявку
			existingRequest, fetchErr := GetLicenseRequestByKey(licenseKey)
			if fetchErr != nil {
				log.Printf("Error fetching existing license request: %v", fetchErr)
				return nil, fetchErr
			}
			if existingRequest != nil {
				log.Printf("Found existing license request: %+v", existingRequest)
				return existingRequest, errors.New("license request with this key already exists")
			}
			log.Printf("License request with key %s exists but could not be fetched.", licenseKey)
			return nil, errors.New("license request with this key already exists, but could not fetch it")
		}
		log.Printf("Error inserting license request: %v", err)
		return nil, err
	}

	log.Printf("Created new license request: %+v", request)
	return &request, nil
}

// Получает заявку на лицензию по license_key.
func GetLicenseRequestByKey(licenseKey string) (*models.LicenseRequest, error) {
	query := `
        SELECT id, status, created_at, license_key
        FROM license_requests
        WHERE license_key = $1;
    `

	var request models.LicenseRequest
	err := DB.QueryRow(query, licenseKey).Scan(
		&request.ID,
		&request.Status,
		&request.CreatedAt,
		&request.LicenseKey,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &request, nil
}

// Получает не одобренные заявки
func GetLicenseRequests() ([]models.LicenseRequest, error) {
	rows, err := DB.Query("SELECT * FROM license_requests WHERE status != 'approved' ORDER BY created_at DESC")
	if err != nil {
		log.Println("Error fetching license requests:", err)
		return nil, err
	}
	defer rows.Close()

	var requests []models.LicenseRequest
	for rows.Next() {
		var req models.LicenseRequest
		if err := rows.Scan(&req.ID, &req.Status, &req.CreatedAt, &req.LicenseKey); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// Одобряет заявку и создает лицензию
func ApproveLicenseRequest(requestID int, tag int) error {
	var licenseKey string
	query := `
        SELECT license_key FROM license_requests
        WHERE id = $1 AND status != 'approved'
        FOR UPDATE;
    `
	err := DB.QueryRow(query, requestID).Scan(&licenseKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("request with ID: %d not found or processed already", requestID)
		}
		return err
	}

	// Генерация подписи
	signature, err := security.SignLicense(licenseKey)
	if err != nil {
		return err
	}

	// Создание лицензии
	licenseQuery := `
        INSERT INTO licenses (license_key, status, tag, license_signature)
        VALUES ($1, 'active', $2, $3)
        RETURNING id;
    `
	var licenseID int
	err = DB.QueryRow(licenseQuery, licenseKey, tag, signature).Scan(&licenseID)
	if err != nil {
		return err
	}

	// Обновление статуса заявки
	updateQuery := `
        UPDATE license_requests
        SET status = 'approved'
        WHERE id = $1;
    `
	_, err = DB.Exec(updateQuery, requestID)
	if err != nil {
		return err
	}

	return nil
}

// Отклоняет заявку
func RejectLicenseRequest(requestID int) error {
	query := `
        UPDATE license_requests
        SET status = 'rejected'
        WHERE id = $1 AND status = 'pending';
    `
	result, err := DB.Exec(query, requestID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("request with ID: %d not found or processed already", requestID)
	}

	return nil
}

// Получает лицензию пользователя по его ID
func GetLicenseByKey(licenseKey string) (*models.License, error) {
	query := `
        SELECT *
        FROM licenses
        WHERE license_key = $1;
    `

	var license models.License
	err := DB.QueryRow(query, licenseKey).Scan(
		&license.ID,
		&license.LicenseKey,
		&license.LicenseSignature,
		&license.Status,
		&license.CreatedAt,
		&license.Tag,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &license, nil
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

// Проверяет, является ли ошибка нарушением уникальности.
func isUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505" // PostgreSQL код ошибки уникального нарушения
}

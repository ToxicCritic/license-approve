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

// CreateLicenseRequest пытается вставить новую заявку, при нарушении уникальности — возвращает уже существующую.
func CreateLicenseRequest(licenseKey string) (*models.LicenseRequest, error) {
	const query = `
		INSERT INTO license_requests (license_key, status)
		VALUES ($1, 'pending')
		RETURNING id, status, created_at, license_key;
	`
	var req models.LicenseRequest
	err := DB.QueryRow(query, licenseKey).Scan(
		&req.ID,
		&req.Status,
		&req.CreatedAt,
		&req.LicenseKey,
	)
	if err != nil {
		if isUniqueViolation(err) {
			log.Printf("License key %s already exists. Fetching existing request.", licenseKey)
			existing, fetchErr := GetLicenseRequestByKey(licenseKey)
			if fetchErr != nil {
				return nil, fetchErr
			}
			if existing != nil {
				return existing, errors.New("license request with this key already exists")
			}
		}
		return nil, err
	}
	log.Printf("Created new license request: %+v", req)
	return &req, nil
}

// GetLicenseRequestByKey возвращает заявку по ключу или nil.
func GetLicenseRequestByKey(licenseKey string) (*models.LicenseRequest, error) {
	const query = `
		SELECT id, status, created_at, license_key
		FROM license_requests
		WHERE license_key = $1;
	`
	var req models.LicenseRequest
	err := DB.QueryRow(query, licenseKey).Scan(
		&req.ID,
		&req.Status,
		&req.CreatedAt,
		&req.LicenseKey,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// GetLicenseRequests возвращает все непроцессированные заявки (status != 'approved').
func GetLicenseRequests() ([]models.LicenseRequest, error) {
	rows, err := DB.Query(`
		SELECT id, status, created_at, license_key
		FROM license_requests
		WHERE status != 'approved'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.LicenseRequest
	for rows.Next() {
		var req models.LicenseRequest
		if err := rows.Scan(&req.ID, &req.Status, &req.CreatedAt, &req.LicenseKey); err != nil {
			return nil, err
		}
		list = append(list, req)
	}
	return list, rows.Err()
}

// GetLicenseRequestsByKey ищет заявки по части ключа (ILIKE).
func GetLicenseRequestsByKey(key string) ([]models.LicenseRequest, error) {
	rows, err := DB.Query(`
		SELECT id, status, created_at, license_key
		FROM license_requests
		WHERE license_key ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
	`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.LicenseRequest
	for rows.Next() {
		var req models.LicenseRequest
		if err := rows.Scan(&req.ID, &req.Status, &req.CreatedAt, &req.LicenseKey); err != nil {
			return nil, err
		}
		list = append(list, req)
	}
	return list, rows.Err()
}

// ApproveLicenseRequest создаёт лицензию по заявке и помечает её как 'approved'.
func ApproveLicenseRequest(requestID, tag int, approver string) error {
	log.Printf("[Approve] request=%d tag=%d approver=%q", requestID, tag, approver)
	if approver == "" || approver == "unknown" {
		return fmt.Errorf("cannot approve license: approver login is undefined")
	}

	var licenseKey string
	const selectQ = `
		SELECT license_key
		  FROM license_requests
		 WHERE id = $1
		   AND status != 'approved'
		  FOR UPDATE
	`
	if err := DB.QueryRow(selectQ, requestID).Scan(&licenseKey); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("request %d not found or already processed", requestID)
		}
		return err
	}

	signature, err := security.SignLicense(licenseKey)
	if err != nil {
		return err
	}

	const insertQ = `
		INSERT INTO licenses
		            (license_key, status, tag, license_signature, approved_by)
		     VALUES ($1, 'active', $2, $3, $4)
	`
	if _, err := DB.Exec(insertQ, licenseKey, tag, signature, approver); err != nil {
		return err
	}

	const updateQ = `
		UPDATE license_requests
		   SET status = 'approved'
		 WHERE id = $1
	`
	_, err = DB.Exec(updateQ, requestID)
	return err
}

// RejectLicenseRequest помечает pending-заявку как rejected.
func RejectLicenseRequest(requestID int) error {
	res, err := DB.Exec(`
		UPDATE license_requests
		SET status = 'rejected'
		WHERE id = $1 AND status = 'pending'
	`, requestID)
	if err != nil {
		return err
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return fmt.Errorf("request %d not found or already processed", requestID)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}
	return false
}

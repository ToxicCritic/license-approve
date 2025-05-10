package db

import (
	"database/sql"
	"server/pkg/models"
)

// GetAllLicenses возвращает список всех лицензий, упорядоченных по дате создания (DESC).
func GetAllLicenses() ([]models.License, error) {
	rows, err := DB.Query(`
		SELECT id, license_key, license_signature, status, created_at, tag
		FROM licenses
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.License
	for rows.Next() {
		var lic models.License
		if err := rows.Scan(
			&lic.ID,
			&lic.LicenseKey,
			&lic.LicenseSignature,
			&lic.Status,
			&lic.CreatedAt,
			&lic.Tag,
		); err != nil {
			return nil, err
		}
		list = append(list, lic)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// GetLicensesByKey ищет лицензии по части ключа (ILIKE).
func GetLicensesByKey(key string) ([]models.License, error) {
	rows, err := DB.Query(`
		SELECT id, license_key, license_signature, status, created_at, tag
		FROM licenses
		WHERE license_key ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
	`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.License
	for rows.Next() {
		var lic models.License
		if err := rows.Scan(
			&lic.ID,
			&lic.LicenseKey,
			&lic.LicenseSignature,
			&lic.Status,
			&lic.CreatedAt,
			&lic.Tag,
		); err != nil {
			return nil, err
		}
		list = append(list, lic)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// GetLicenseByKey возвращает лицензию по точному совпадению ключа или nil.
func GetLicenseByKey(licenseKey string) (*models.License, error) {
	const query = `
		SELECT id, license_key, license_signature, status, created_at, tag
		FROM licenses
		WHERE license_key = $1
	`
	var lic models.License
	err := DB.QueryRow(query, licenseKey).Scan(
		&lic.ID,
		&lic.LicenseKey,
		&lic.LicenseSignature,
		&lic.Status,
		&lic.CreatedAt,
		&lic.Tag,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &lic, nil
}

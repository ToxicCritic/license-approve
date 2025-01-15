package db

import (
	"LicenseApp/pkg/licensegen"
	"LicenseApp/pkg/models"
	"LicenseApp/pkg/security"
	"fmt"
	"log"
)

func GetLicenseRequests() ([]models.LicenseRequest, error) {
	rows, err := DB.Query("SELECT * FROM license_requests")
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

func RejectLicenseRequest(id int) error {
	_, err := DB.Exec("UPDATE license_requests SET status = 'rejected' WHERE id = $1", id)
	if err != nil {
		log.Println("Error rejecting license request:", err)
		return err
	}
	return nil
}

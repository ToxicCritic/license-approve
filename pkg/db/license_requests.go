package db

import (
	"LicenseApp/pkg/models"
	"log"
)

// Получить список всех запросов на лицензии
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

// Одобрить запрос на лицензию
func ApproveLicenseRequest(id int) error {
	_, err := DB.Exec("UPDATE license_requests SET status = 'approved' WHERE id = $1", id)
	if err != nil {
		log.Println("Error updating license request:", err)
		return err
	}
	return nil
}

// Отклонить запрос на лицензию
func RejectLicenseRequest(id int) error {
	_, err := DB.Exec("UPDATE license_requests SET status = 'rejected' WHERE id = $1", id)
	if err != nil {
		log.Println("Error rejecting license request:", err)
		return err
	}
	return nil
}

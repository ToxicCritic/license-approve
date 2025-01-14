package handlers

import (
	"LicenseApp/pkg/db"
	"encoding/json"
	"net/http"
	"strconv"
)

// Обработчик для получения запросов на лицензии
func GetLicenseRequestsHandler(w http.ResponseWriter, r *http.Request) {
	requests, err := db.GetLicenseRequests()
	if err != nil {
		http.Error(w, "Failed to get license requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// Обработчик для одобрения запроса на лицензию
func ApproveLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid license request ID", http.StatusBadRequest)
		return
	}

	err = db.ApproveLicenseRequest(id)
	if err != nil {
		http.Error(w, "Failed to approve license request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("License request approved"))
}

// Обработчик для отклонения запроса на лицензию
func RejectLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid license request ID", http.StatusBadRequest)
		return
	}

	err = db.RejectLicenseRequest(id)
	if err != nil {
		http.Error(w, "Failed to reject license request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("License request rejected"))
}

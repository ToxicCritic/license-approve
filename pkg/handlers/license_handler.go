package handlers

import (
	"LicenseApp/pkg/db"
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestID struct {
	ID int `json:"id"`
}

func CreateLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	type requestBody struct {
		UserID    int    `json:"user_id"`
		PublicKey string `json:"public_key"`
	}

	var input requestBody
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	newID, err := db.CreateLicenseRequest(input.UserID, input.PublicKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create license request: %v", err), http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		RequestID int    `json:"request_id"`
		Status    string `json:"status"`
		Message   string `json:"message"`
	}

	resp := responseBody{
		RequestID: newID,
		Status:    "pending",
		Message:   "License request created successfully",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Обработчик для получения запросов на лицензии
func GetLicenseRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method, use GET", http.StatusMethodNotAllowed)
		return
	}

	requests, err := db.GetLicenseRequests()
	if err != nil {
		http.Error(w, "Failed to get license requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// Обработчик для подтверждения запроса на лицензию
func ApproveLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, use POST", http.StatusMethodNotAllowed)
		return
	}

	var req RequestID
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON body", http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	err := db.ApproveLicenseRequest(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to approve license request: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("License request approved"))
}

// Обработчик для отклонения запроса на лицензию
func RejectLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, use POST", http.StatusMethodNotAllowed)
		return
	}

	var req RequestID
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse JSON body", http.StatusBadRequest)
		return
	}
	if req.ID <= 0 {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	err := db.RejectLicenseRequest(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reject license request: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("License request rejected"))
}

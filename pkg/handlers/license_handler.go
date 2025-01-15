package handlers

import (
	"LicenseApp/internal/templates"
	"LicenseApp/pkg/db"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type RequestID struct {
	ID int `json:"id"`
}

// Обработчик для создания заявки на лицензию
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

	tmpl := templates.ParseTemplates()

	err = tmpl.ExecuteTemplate(w, "admin_requests.html", requests)
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func ApproveLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, use POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID (not an integer)", http.StatusBadRequest)
		return
	}

	err = db.ApproveLicenseRequest(requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to approve license request: %v", err), http.StatusInternalServerError)
		return
	}

	// После успеха редиректим обратно на список
	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

// Обработчик для отклонения запроса на лицензию
func RejectLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, use POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID (not an integer)", http.StatusBadRequest)
		return
	}

	err = db.RejectLicenseRequest(requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reject license request: %v", err), http.StatusInternalServerError)
		return
	}

	// После успеха редиректим обратно на список
	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/pkg/db"
	"server/pkg/models"
	"server/templates"
	"strconv"
)

// CreateLicenseRequestPayload описывает JSON-запрос клиента.
type CreateLicenseRequestPayload struct {
	LicenseKey string `json:"license_key"`
}

// CreateLicenseRequestResponse описывает JSON-ответ.
type CreateLicenseRequestResponse struct {
	Message    string `json:"message"`
	RequestID  int    `json:"request_id"`
	LicenseKey string `json:"license_key,omitempty"`
}

// CreateLicenseRequestHandler обрабатывает POST /api/create-license-request.
func CreateLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var payload CreateLicenseRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding payload: %v", err)
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}
	if payload.LicenseKey == "" {
		http.Error(w, "Missing license_key.", http.StatusBadRequest)
		return
	}

	req, err := db.CreateLicenseRequest(payload.LicenseKey)
	if err != nil {
		if err.Error() == "license request with this key already exists" && req != nil {
			resp := CreateLicenseRequestResponse{
				Message:   "A pending license request already exists.",
				RequestID: req.ID,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
			return
		}
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	resp := CreateLicenseRequestResponse{
		Message:    "License request created successfully.",
		RequestID:  req.ID,
		LicenseKey: req.LicenseKey,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetLicenseRequestsHandler обрабатывает GET /admin/license-requests и поиск по ?q=.
func GetLicenseRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	var list []models.LicenseRequest
	var err error
	if q != "" {
		list, err = db.GetLicenseRequestsByKey(q)
	} else {
		list, err = db.GetLicenseRequests()
	}
	if err != nil {
		log.Printf("Error fetching requests: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	tmpl := templates.ParseTemplates()
	data := struct {
		Requests []models.LicenseRequest
		Query    string
	}{list, q}

	if err := tmpl.ExecuteTemplate(w, "admin_requests.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// ApproveLicenseRequestHandler обрабатывает POST /admin/approve-license.
func ApproveLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", http.StatusBadRequest)
		return
	}
	idStr, tagStr := r.FormValue("id"), r.FormValue("tag")
	if idStr == "" || tagStr == "" {
		http.Error(w, "Missing id or tag", http.StatusBadRequest)
		return
	}
	id, err1 := strconv.Atoi(idStr)
	tag, err2 := strconv.Atoi(tagStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid id or tag", http.StatusBadRequest)
		return
	}
	if err := db.ApproveLicenseRequest(id, tag); err != nil {
		log.Printf("Error approving: %v", err)
		http.Error(w, fmt.Sprintf("Failed to approve: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

// RejectLicenseRequestHandler обрабатывает POST /admin/reject-license.
func RejectLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", http.StatusBadRequest)
		return
	}
	idStr := r.FormValue("id")
	if idStr == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := db.RejectLicenseRequest(id); err != nil {
		log.Printf("Error rejecting: %v", err)
		http.Error(w, fmt.Sprintf("Failed to reject: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

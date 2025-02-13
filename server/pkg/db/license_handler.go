package db

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"example.com/licence-approval/server/templates"
)

// Представляет структуру запроса для создания заявки на лицензию
type CreateLicenseRequestPayload struct {
	LicenseKey string `json:"license_key"`
}

// Представляет структуру ответа на создание заявки
type CreateLicenseRequestResponse struct {
	Message    string `json:"message"`
	RequestID  int    `json:"request_id"`
	LicenseKey string `json:"license_key,omitempty"`
}

// Обрабатывает создание новой заявки на лицензию
func CreateLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var payload CreateLicenseRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding license request payload: %v", err)
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}

	if payload.LicenseKey == "" {
		http.Error(w, "Missing license_key.", http.StatusBadRequest)
		return
	}

	// Создание новой заявки на лицензию с предоставленным license_key
	request, err := CreateLicenseRequest(payload.LicenseKey)
	if err != nil {
		// Проверка, если заявка уже существует
		if err.Error() == "license request with this key already exists" && request != nil {
			response := CreateLicenseRequestResponse{
				Message:   "A pending license request already exists.",
				RequestID: request.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // HTTP 409 Conflict
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Error sending existing request response: %v", err)
			}
			return
		}

		log.Printf("Error creating license request: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	// Формирование ответа с информацией о созданной заявке
	response := CreateLicenseRequestResponse{
		Message:    "License request created successfully.",
		RequestID:  request.ID,
		LicenseKey: request.LicenseKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // HTTP 201 Created
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// Обрабатывает получение всех заявок на лицензии для административной панели.
func GetLicenseRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method, use GET", http.StatusMethodNotAllowed)
		return
	}

	requests, err := GetLicenseRequests()
	if err != nil {
		log.Printf("Error fetching license requests: %v", err)
		http.Error(w, "Failed to get license requests", http.StatusInternalServerError)
		return
	}

	// Парсинг и выполнение шаблона для отображения заявок
	tmpl := templates.ParseTemplates()

	err = tmpl.ExecuteTemplate(w, "admin_requests.html", requests)
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// Обрабатывает одобрение заявки на лицензию.
func ApproveLicenseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, use POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Извлечение параметров из формы
	idStr := r.FormValue("id")
	tagStr := r.FormValue("tag")
	if idStr == "" || tagStr == "" {
		http.Error(w, "Missing request ID or tag", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid request ID (not an integer)", http.StatusBadRequest)
		return
	}

	tag, err := strconv.Atoi(tagStr)
	if err != nil || tag <= 0 || tag > 1000000 {
		http.Error(w, "Invalid TAG value (must be a positive integer between 1 and 1000000)", http.StatusBadRequest)
		return
	}

	// Одобрение заявки на лицензию
	err = ApproveLicenseRequest(requestID, tag)
	if err != nil {
		log.Printf("Error approving license request ID %d: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Failed to approve license request: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

// Ообрабатывает отклонение заявки на лицензию.
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

	err = RejectLicenseRequest(requestID)
	if err != nil {
		log.Printf("Error rejecting license request ID %d: %v", requestID, err)
		http.Error(w, fmt.Sprintf("Failed to reject license request: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

// Проверяет статус лицензии по license_key.
func CheckLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET.", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение license_key из параметров запроса
	licenseKey := r.URL.Query().Get("license_key")
	if licenseKey == "" {
		http.Error(w, "Missing license_key parameter.", http.StatusBadRequest)
		return
	}

	// Получение информации о лицензии по license_key
	license, err := GetLicenseByKey(licenseKey)
	if err != nil {
		log.Printf("Error fetching license for key %s: %v", licenseKey, err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	var response struct {
		HasLicense bool   `json:"has_license"`
		Message    string `json:"message"`
	}

	if license != nil {
		// Лицензия найдена, формирование ответа на основе статуса
		switch license.Status {
		case "active":
			response.HasLicense = true
			response.Message = fmt.Sprintf("License is active. TAG: %v", license.Tag)
		case "revoked", "expired":
			response.HasLicense = false
			response.Message = "License is not active."
		default:
			response.HasLicense = false
			response.Message = "License is not active."
		}
	} else {
		// Если лицензия не найдена, проверка на наличие ожидающей заявки
		pendingRequest, err := GetLicenseRequestByKey(licenseKey)
		if err != nil {
			log.Printf("Error fetching license request for key %s: %v", licenseKey, err)
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
			return
		}

		if pendingRequest != nil {
			// Существует заявка, проверяем её статус
			switch pendingRequest.Status {
			case "pending":
				response.HasLicense = false
				response.Message = "License request is pending."
			case "rejected":
				response.HasLicense = false
				response.Message = "License request has been rejected."
			default:
				response.HasLicense = false
				response.Message = "License request is not active."
			}
		} else {
			// Лицензия не активна и нет заявок
			response.HasLicense = false
			response.Message = "License is not active."
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding license status response: %v", err)
	}
}

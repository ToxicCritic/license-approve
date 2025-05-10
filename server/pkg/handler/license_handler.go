package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/pkg/db"
	"server/pkg/models"
	"server/templates"
)

// CheckLicenseHandler обрабатывает GET /api/check-license
func CheckLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}
	licenseKey := r.URL.Query().Get("license_key")
	if licenseKey == "" {
		http.Error(w, "Missing license_key", http.StatusBadRequest)
		return
	}
	lic, err := db.GetLicenseByKey(licenseKey)
	if err != nil {
		log.Printf("Error fetching license: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var resp struct {
		HasLicense bool   `json:"has_license"`
		Message    string `json:"message"`
	}

	if lic != nil {
		switch lic.Status {
		case "active":
			resp.HasLicense = true
			resp.Message = fmt.Sprintf("License is active. TAG: %v", lic.Tag)
		default:
			resp.HasLicense = false
			resp.Message = "License is not active."
		}
	} else {
		pending, err := db.GetLicenseRequestByKey(licenseKey)
		if err != nil {
			log.Printf("Error fetching request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if pending != nil {
			resp.HasLicense = false
			resp.Message = fmt.Sprintf("License request is %s.", pending.Status)
		} else {
			resp.HasLicense = false
			resp.Message = "License is not active."
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLicensesHandler обрабатывает GET /admin/licenses и поиск по ?q=
func GetLicensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Use GET", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	var list []models.License
	var err error
	if q != "" {
		list, err = db.GetLicensesByKey(q)
	} else {
		list, err = db.GetAllLicenses()
	}
	if err != nil {
		log.Printf("Error fetching licenses: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl := templates.ParseTemplates()
	data := struct {
		Licenses []models.License
		Query    string
	}{list, q}

	if err := tmpl.ExecuteTemplate(w, "admin_licenses.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

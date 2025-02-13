package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"example.com/licence-approval/client/pkg/errors"
)

// Представляет структуру ответа от сервера на запрос проверки лицензии.
type CheckLicenseResponse struct {
	HasLicense bool   `json:"has_license"`
	Message    string `json:"message"`
}

func CheckLicense(client *http.Client, serverURL, licenseKey string) (bool, string, error) {
	// Формирование URL с параметром license_key
	url := fmt.Sprintf("%s/api/check-license?license_key=%s", serverURL, licenseKey)

	// Создание GET-запроса
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Отправка запроса
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Чтение тела ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Проверка HTTP статуса
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("server returned non-OK status: %s, body: %s", resp.Status, string(body))
	}

	// Декодирование JSON-ответа
	var checkResp CheckLicenseResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return false, "", fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if checkResp.HasLicense {
		return true, checkResp.Message, nil
	}

	// Обработка оставшихся сообщений
	switch checkResp.Message {
	case "License request is pending.":
		return false, checkResp.Message, nil
	case "License request has been rejected.":
		return false, checkResp.Message, &errors.LicenseRejectedError{Status: checkResp.Message}
	default:
		return false, checkResp.Message, nil
	}
}

// Представляет структуру запроса для создания заявки на лицензию.
type CreateLicenseRequestPayload struct {
	LicenseKey string `json:"license_key"`
}

// Представляет структуру ответа на создание заявки
type CreateLicenseRequestResponse struct {
	Message    string `json:"message"`
	RequestID  int    `json:"request_id"`
	LicenseKey string `json:"license_key,omitempty"`
}

// Отправляет POST-запрос на сервер для создания заявки на лицензию.
func CreateLicenseRequest(client *http.Client, serverURL, licenseKey string) (int, error) {
	url := fmt.Sprintf("%s/api/create-license-request", serverURL)

	// Подготовка тела запроса с license_key
	payload := CreateLicenseRequestPayload{
		LicenseKey: licenseKey,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	// Обработка различных статусов ответа
	if resp.StatusCode == http.StatusCreated {
		// Успешное создание заявки
		var response CreateLicenseRequestResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return 0, fmt.Errorf("failed to parse JSON response: %v", err)
		}
		return response.RequestID, nil
	} else if resp.StatusCode == http.StatusConflict {
		// Заявка уже существует
		var response struct {
			Message   string `json:"message"`
			RequestID int    `json:"request_id"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0, fmt.Errorf("failed to parse JSON conflict response: %v", err)
		}
		return response.RequestID, &errors.LicenseRequestExistsError{RequestID: response.RequestID}
	} else {
		// Другие ошибки
		return 0, fmt.Errorf("failed to create license request: status %s, body: %s", resp.Status, string(body))
	}
}

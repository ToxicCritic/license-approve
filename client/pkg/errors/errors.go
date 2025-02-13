package errors

import (
	"fmt"
)

// Представляет ошибку, когда заявка уже существует.
type LicenseRequestExistsError struct {
	RequestID int
}

func (e *LicenseRequestExistsError) Error() string {
	return fmt.Sprintf("license request already exists with ID %d", e.RequestID)
}

// Представляет ошибку, связанную с определённым статусом лицензии.
type LicenseStatusError struct {
	Status string
}

func (e *LicenseStatusError) Error() string {
	return fmt.Sprintf("license status: %s", e.Status)
}

// Представляет ошибку, связанную с отклонением заявки на лицензию
type LicenseRejectedError struct {
	Status string
}

func (e *LicenseRejectedError) Error() string {
	return fmt.Sprintf("license request rejected: %s", e.Status)
}

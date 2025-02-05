// internal/api/v1/httpapi/handler/user.go
package handler

import (
	"net/http"
)

// Аналогично
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}

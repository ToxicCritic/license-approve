// internal/api/v1/httpapi/handler/group.go
package handler

import (
	"net/http"
)

// Имитируем endpoins:
//
//	GET /api/v1/config/groups
//	POST /api/v1/config/group
//	...
//
// Упрощённо вернём "501 Not Implemented"
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented (mock)", http.StatusNotImplemented)
}

// И т.д.

// internal/api/v1/httpapi/router/router.go
package router

import (
	"mock-authserver/internal/api/v1/httpapi/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	// Группы
	r.Get("/api/v1/config/groups", h.GetGroups)
	r.Post("/api/v1/config/group", h.CreateGroup)
	// и т.д.

	// Пользователи
	r.Get("/api/v1/config/users", h.GetAllUsers)
	r.Post("/api/v1/config/user", h.CreateUser)
	r.Put("/api/v1/config/user/{UserLogin}", h.UpdateUser)
	r.Delete("/api/v1/config/user/{UserLogin}", h.DeleteUser)

	// OAuth2.0 (псевдо)
	r.Post("/oauth/token", h.TokenHandler)

	// Допустим, для "авторизации" (старт): GET /auth/login
	r.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Mock login page", http.StatusOK)
	})

	return r
}

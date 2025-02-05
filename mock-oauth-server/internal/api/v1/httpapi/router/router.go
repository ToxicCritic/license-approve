package router

import (
	"mock-oauth-server/internal/api/v1/httpapi/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// New создаёт chi.Router и регистрирует все эндпоинты
func New() http.Handler {
	r := chi.NewRouter()

	h := handler.NewHandler()

	r.Get("/api/v1/config/groups", h.GetGroups)
	r.Post("/api/v1/config/group", h.CreateGroup)

	r.Get("/api/v1/config/users", h.GetAllUsers)
	r.Post("/api/v1/config/user", h.CreateUser)
	r.Put("/api/v1/config/user/{UserLogin}", h.UpdateUser)
	r.Delete("/api/v1/config/user/{UserLogin}", h.DeleteUser)

	// OAuth2
	r.Get("/authorize", h.AuthorizeHandler)
	r.Post("/authorize", h.AuthorizeHandler)
	r.Post("/token", h.TokenHandler)

	return r
}

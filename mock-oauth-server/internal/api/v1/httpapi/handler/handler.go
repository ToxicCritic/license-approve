// internal/api/v1/httpapi/handler/handler.go
package handler

import (
	"mock-authserver/internal/repository/inmem"
	"mock-authserver/internal/service/jwtaccess"
	"mock-authserver/internal/usecase"
)

type Handler struct {
	store   *inmem.Store
	usecase *usecase.Usecase
	jwtSvc  *jwtaccess.JWTService
}

func New(store *inmem.Store, uc *usecase.Usecase, js *jwtaccess.JWTService) *Handler {
	return &Handler{
		store:   store,
		usecase: uc,
		jwtSvc:  js,
	}
}

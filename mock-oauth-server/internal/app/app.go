// internal/app/app.go
package app

import (
	"mock-authserver/config"
	"mock-authserver/internal/api/v1/httpapi/handler"
	"mock-authserver/internal/api/v1/httpapi/router"
	"mock-authserver/internal/repository/inmem"
	"mock-authserver/internal/service/jwtaccess"
	"mock-authserver/internal/usecase"
	"net/http"
)

// App хранит роутер (как минимум).
type App struct {
	Router http.Handler
}

func New(cfg *config.Config) (*App, error) {
	// In-memory store
	store := inmem.NewStore()

	// UseCase
	uc := usecase.New(store)

	// JWT service
	jwtSvc := jwtaccess.NewJWTService(store, cfg.SecretKey)

	// Handler
	h := handler.New(store, uc, jwtSvc)

	// Router
	r := router.New(h)

	return &App{
		Router: r,
	}, nil
}

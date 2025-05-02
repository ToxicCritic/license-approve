package handler

import (
	"database/sql"
	"mock-oauth-server/internal/repository/inmem"
	"mock-oauth-server/internal/usecase"
)

type Handler struct {
	usecases *usecase.Usecases
	store    *inmem.Store
}

func NewHandler(db *sql.DB) *Handler {
	store := inmem.NewStore(db)
	uc := usecase.NewUsecases(store)
	return &Handler{
		store:    store,
		usecases: uc,
	}
}

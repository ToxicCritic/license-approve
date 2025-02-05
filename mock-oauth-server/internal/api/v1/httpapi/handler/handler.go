package handler

import (
	"mock-oauth-server/internal/repository/inmem"
	"mock-oauth-server/internal/usecase"
)

type Handler struct {
	usecases *usecase.Usecases
	store    *inmem.Store
}

func NewHandler() *Handler {
	store := inmem.NewStore()        // хранит users, groups, tokens
	uc := usecase.NewUsecases(store) // бизнес-логика
	return &Handler{
		store:    store,
		usecases: uc,
	}
}

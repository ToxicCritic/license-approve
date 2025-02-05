// internal/usecase/usecase.go
package usecase

import (
	"mock-authserver/internal/entity"
	"mock-authserver/internal/repository/inmem"
)

type Usecase struct {
	store *inmem.Store
}

func New(store *inmem.Store) *Usecase {
	return &Usecase{store: store}
}

// Можно добавить логику типа "создать пользователя", "создать группу" и т.д.
// Для упрощённого мок-сервера можно ограничиться минимальным набором.
func (u *Usecase) ValidateUser(username, password string) (*entity.User, bool) {
	return u.store.FindUser(username, password)
}

package usecase

import (
	"errors"
	"mock-oauth-server/internal/repository/inmem"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Usecases struct {
	store *inmem.Store
}

func NewUsecases(store *inmem.Store) *Usecases {
	return &Usecases{
		store: store,
	}
}

func (u *Usecases) GetAllUsers() []map[string]string {
	var list []map[string]string
	for _, user := range u.store.Users {
		list = append(list, map[string]string{
			"username":   user.Username,
			"created_at": user.CreatedAt.Format(time.RFC3339),
		})
	}
	return list
}

func (u *Usecases) CreateUser(username, password string) error {
	if _, exists := u.store.Users[username]; exists {
		return errors.New("user already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.store.Users[username] = &inmem.User{
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	return nil
}

func (u *Usecases) GetAllGroups() []string {
	var list []string
	for name := range u.store.Groups {
		list = append(list, name)
	}
	return list
}

func (u *Usecases) CreateGroup(name string) error {
	if _, ok := u.store.Groups[name]; ok {
		return errors.New("group already exists")
	}
	u.store.Groups[name] = &inmem.Group{Name: name}
	return nil
}

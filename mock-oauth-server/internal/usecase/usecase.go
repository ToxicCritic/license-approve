package usecase

import (
	"errors"
	"mock-oauth-server/internal/repository/inmem"
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
	var all []map[string]string
	for _, user := range u.store.Users {
		all = append(all, map[string]string{
			"id":       user.ID,
			"username": user.Username,
		})
	}
	return all
}

func (u *Usecases) CreateUser(username, password string) error {
	for _, v := range u.store.Users {
		if v.Username == username {
			return errors.New("user already exists")
		}
	}
	newID := "user-x" // generate
	u.store.Users[newID] = &inmem.User{
		ID:       newID,
		Username: username,
		Password: password,
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

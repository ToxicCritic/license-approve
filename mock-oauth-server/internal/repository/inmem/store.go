// internal/repository/inmem/store.go
package inmem

import (
	"mock-authserver/internal/entity"
	"sync"
	"time"
)

type Store struct {
	mu            sync.Mutex
	users         map[string]*entity.User
	clients       map[string]*entity.Client
	accessTokens  map[string]*entity.AccessToken
	refreshTokens map[string]*entity.RefreshToken
}

func NewStore() *Store {
	s := &Store{
		users:         make(map[string]*entity.User),
		clients:       make(map[string]*entity.Client),
		accessTokens:  make(map[string]*entity.AccessToken),
		refreshTokens: make(map[string]*entity.RefreshToken),
	}
	// Пример: добавим одного пользователя "admin" / "password"
	s.users["u1"] = &entity.User{ID: "u1", Username: "admin", Password: "password"}

	// Пример: один клиент
	s.clients["mock-client-id"] = &entity.Client{
		ID:     "mock-client-id",
		Secret: "mock-client-secret",
		Public: false,
		UserID: "u1",
	}
	return s
}

// ---- Методы для пользователей ----
func (s *Store) FindUser(username, password string) (*entity.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.Username == username && user.Password == password {
			return user, true
		}
	}
	return nil, false
}

// ---- Методы для клиентов ----
func (s *Store) FindClient(clientID, clientSecret string) (*entity.Client, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cli, ok := s.clients[clientID]
	if !ok {
		return nil, false
	}
	if cli.Secret != clientSecret {
		return nil, false
	}
	return cli, true
}

// ---- Методы для access_tokens ----
func (s *Store) SaveAccessToken(at *entity.AccessToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessTokens[at.Token] = at
}

func (s *Store) GetAccessToken(token string) (*entity.AccessToken, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	at, ok := s.accessTokens[token]
	return at, ok
}

// ---- Методы для refresh_tokens ----
func (s *Store) SaveRefreshToken(rt *entity.RefreshToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[rt.Token] = rt
}

func (s *Store) GetRefreshToken(token string) (*entity.RefreshToken, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, ok := s.refreshTokens[token]
	return rt, ok
}

func (s *Store) DeleteRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, token)
}

// Удалим просроченные
func (s *Store) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for tk, at := range s.accessTokens {
		if now.After(at.Expiry) {
			delete(s.accessTokens, tk)
		}
	}
	for tk, rt := range s.refreshTokens {
		if now.After(rt.Expiry) {
			delete(s.refreshTokens, tk)
		}
	}
}

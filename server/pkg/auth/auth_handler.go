package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"server/config"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// init регистрирует типы для кодирования сессий (oauth2.Token, UserSession)
func init() {
	gob.Register(&oauth2.Token{})
	gob.Register(&UserSession{})
}

// UserSession хранит токены и время истечения для сессии пользователя
type UserSession struct {
	Username     string
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

// Store глобальное хранилище cookie-сессий
var Store *sessions.CookieStore

// SetupSessionStore инициализирует хранилище сессий с использованием SESSION_SECRET из конфигурации
func SetupSessionStore(cfg *config.Config) {
	sessionSecret := cfg.SessionSecret
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET не установлен в конфигурации")
	}
	Store = sessions.NewCookieStore([]byte(sessionSecret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true, // prod
	}
}

// generateStateToken генерирует случайный токен для защиты от CSRF
func generateStateToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Ошибка генерации состояния:", err)
		return "default-state"
	}
	return base64.URLEncoding.EncodeToString(b)
}

// LoginHandler запускает процесс OAuth2 авторизации, генерируя state и перенаправляя пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Генерация state-токена для защиты от CSRF атак
	state := generateStateToken()

	session, err := Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	url := OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CallbackHandler обрабатывает обратный вызов OAuth2: проверяет state, обменивает код на токен и сохраняет сессию
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	sess, err := Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if sess.Values["state"] != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	token, err := ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	username, _ := token.Extra("username").(string)
	if username == "" {
		username, _ = token.Extra("login").(string)
	}

	userSession := &UserSession{
		Username:     username,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}

	sess.Values["authenticated"] = true
	sess.Values["user"] = userSession
	if err := sess.Save(r, w); err != nil {
		http.Error(w, "Session save failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/license-requests", http.StatusSeeOther)
}

// LogoutHandler завершает сессию пользователя, очищая данные и удаляя куки, затем перенаправляет на главную страницу
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["authenticated"] = false
	session.Values["user"] = nil
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

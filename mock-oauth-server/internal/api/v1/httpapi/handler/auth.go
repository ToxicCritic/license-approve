package handler

import (
	"fmt"
	"log"
	"mock-oauth-server/internal/repository/inmem"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthorizeHandler обрабатывает получение authorization code
func (h *Handler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.handleAuthorizeGet(w, r)
	} else if r.Method == http.MethodPost {
		h.handleAuthorizePost(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleAuthorizeGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	responseType := q.Get("response_type")
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	scope := q.Get("scope")

	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if clientID == "" || redirectURI == "" {
		http.Error(w, "missing client_id or redirect_uri", http.StatusBadRequest)
		return
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Mock OAuth2.0 • Authorize</title>
		<!-- Bootstrap 5 -->
		<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
		<style>
			body {
				background: #f8f9fa;
				display: flex;
				align-items: center;
				justify-content: center;
				height: 100vh;
				font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
			}
			.auth-card {
				width: 360px;
				padding: 2rem 2.5rem;
				border-radius: 12px;
				background: #ffffff;
				box-shadow: 0 4px 12px rgba(0,0,0,.1);
			}
			.auth-card h3 {
				margin-bottom: 1.5rem;
			}
		</style>
	</head>
	<body>
		<div class="auth-card">
			<h3 class="text-center mb-4">Авторизация</h3>
			<form method="POST">
				<input type="hidden" name="response_type" value="%s">
				<input type="hidden" name="client_id"     value="%s">
				<input type="hidden" name="redirect_uri"  value="%s">
				<input type="hidden" name="state"         value="%s">
				<input type="hidden" name="scope"         value="%s">
	
				<div class="mb-3">
					<label class="form-label">Логин</label>
					<input type="text" class="form-control" name="username" required>
				</div>
				<div class="mb-3">
					<label class="form-label">Пароль</label>
					<input type="password" class="form-control" name="password" required>
				</div>
				<button type="submit" class="btn btn-primary w-100" style="background-color: #198754; border-color: #198754;">Войти</button>
			</form>
		</div>
	
		<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
	</body>
	</html>
	`, responseType, clientID, redirectURI, state, scope)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *Handler) handleAuthorizePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	// Сохраняем входные параметры для редиректа
	responseType := r.FormValue("response_type")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	scope := r.FormValue("scope")

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Проверяем учётку через Store
	user, ok := h.store.Users[username]
	if !ok || bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) != nil {
		params := url.Values{}
		params.Set("response_type", responseType)
		params.Set("client_id", clientID)
		params.Set("redirect_uri", redirectURI)
		params.Set("state", state)
		params.Set("scope", scope)
		params.Set("error", "invalid_credentials")
		http.Redirect(w, r, "/authorize?"+params.Encode(), http.StatusFound)
		return
	}

	if responseType != "code" {
		params := url.Values{
			"response_type": {responseType},
			"client_id":     {clientID},
			"redirect_uri":  {redirectURI},
			"state":         {state},
			"scope":         {scope},
			"error":         {"unsupported_response_type"},
		}
		http.Redirect(w, r, "/authorize?"+params.Encode(), http.StatusFound)
		return
	}

	// Генерируем код и сохраняем
	code := h.store.GenerateRandomString(32)
	h.store.AuthorizationCodes[code] = &inmem.AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		UserID:      user.Username,
		RedirectURI: redirectURI,
		Expiry:      time.Now().Add(5 * time.Minute),
	}

	// Редиректим на redirect_uri с code и state
	u, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect URI", http.StatusBadRequest)
		return
	}
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	log.Printf("[AuthorizePost] code=%s -> redirect to %s\n", code, u.String())
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// mock-oauth-server/handlers.go

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// In-memory хранилища
var (
	users              = map[string]*User{}
	authorizationCodes = map[string]*AuthorizationCode{}
	accessTokens       = map[string]*AccessToken{}
	refreshTokens      = map[string]*RefreshToken{}
)

// init инициализирует начальные данные
func init() {
	// Добавляем одного пользователя
	users["admin"] = &User{
		ID:       "user-1",
		Username: "admin",
		Password: "password",
	}
}

// authorizeHandler обрабатывает запросы на авторизацию
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметры запроса
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")

	// Проверяем тип ответа
	if responseType != "code" {
		http.Error(w, "Unsupported response type", http.StatusBadRequest)
		return
	}

	// Проверяем клиента
	if !validateClient(clientID, "") { // Здесь не проверяем секрет клиента для авторизации
		http.Error(w, "Invalid client ID", http.StatusUnauthorized)
		return
	}

	// Отображаем форму авторизации
	if r.Method == "GET" {
		// Отображение простой HTML-формы
		html := fmt.Sprintf(`
			<html>
			<body>
				<h2>Mock OAuth2.0 Authorization</h2>
				<form method="POST" action="/authorize">
					<input type="hidden" name="response_type" value="%s" />
					<input type="hidden" name="client_id" value="%s" />
					<input type="hidden" name="redirect_uri" value="%s" />
					<input type="hidden" name="state" value="%s" />
					<input type="hidden" name="scope" value="%s" />
					<label>Username:</label><input type="text" name="username" /><br/>
					<label>Password:</label><input type="password" name="password" /><br/>
					<input type="submit" value="Authorize" />
				</form>
			</body>
			</html>
		`, responseType, clientID, redirectURI, state, scope)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	// Обработка POST-запроса авторизации
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Аутентификация пользователя
		user, err := findUser(username, password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Генерация авторизационного кода
		code, err := generateRandomString(32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Сохранение кода в хранилище
		authorizationCodes[code] = &AuthorizationCode{
			Code:        code,
			ClientID:    clientID,
			UserID:      user.ID,
			RedirectURI: redirectURI,
			Expiry:      time.Now().Add(10 * time.Minute),
		}

		// Перенаправление на redirect_uri с кодом и состоянием
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, r.FormValue("state"))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// tokenHandler обрабатывает запросы на получение токенов
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим form данные
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant обрабатывает grant_type=authorization_code
func handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Проверка клиента
	if !validateClient(clientID, clientSecret) {
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// Проверка авторизационного кода
	authCode, exists := authorizationCodes[code]
	if !exists || authCode.Expiry.Before(time.Now()) || authCode.ClientID != clientID || authCode.RedirectURI != redirectURI {
		http.Error(w, "Invalid or expired authorization code", http.StatusBadRequest)
		return
	}

	// Удаление использованного кода
	delete(authorizationCodes, code)

	// Генерация токенов
	accessToken, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	refreshToken, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Сохранение токенов
	accessTokens[accessToken] = &AccessToken{
		Token:     accessToken,
		UserID:    authCode.UserID,
		ClientID:  clientID,
		Expiry:    time.Now().Add(1 * time.Hour),
		TokenType: "Bearer",
	}
	refreshTokens[refreshToken] = &RefreshToken{
		Token:    refreshToken,
		UserID:   authCode.UserID,
		ClientID: clientID,
		Expiry:   time.Now().Add(24 * time.Hour),
	}

	// Формирование ответа
	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleRefreshTokenGrant обрабатывает grant_type=refresh_token
func handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Проверка клиента
	if !validateClient(clientID, clientSecret) {
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// Проверка refresh token
	refToken, exists := refreshTokens[refreshToken]
	if !exists || refToken.Expiry.Before(time.Now()) || refToken.ClientID != clientID {
		http.Error(w, "Invalid or expired refresh token", http.StatusBadRequest)
		return
	}

	// Генерация новых токенов
	newAccessToken, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Сохранение новых токенов
	accessTokens[newAccessToken] = &AccessToken{
		Token:     newAccessToken,
		UserID:    refToken.UserID,
		ClientID:  clientID,
		Expiry:    time.Now().Add(1 * time.Hour),
		TokenType: "Bearer",
	}
	refreshTokens[newRefreshToken] = &RefreshToken{
		Token:    newRefreshToken,
		UserID:   refToken.UserID,
		ClientID: clientID,
		Expiry:   time.Now().Add(24 * time.Hour),
	}

	// Удаление старого refresh token
	delete(refreshTokens, refreshToken)

	// Формирование ответа
	resp := map[string]interface{}{
		"access_token":  newAccessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": newRefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// userInfoHandler возвращает информацию о пользователе на основе access token
func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Извлечение Bearer токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	var accessToken string
	fmt.Sscanf(authHeader, "Bearer %s", &accessToken)

	// Проверка токена
	token, exists := accessTokens[accessToken]
	if !exists || token.Expiry.Before(time.Now()) {
		http.Error(w, "Invalid or expired access token", http.StatusUnauthorized)
		return
	}

	// Получение пользователя
	user, exists := users[token.UserID]
	if !exists {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Формирование ответа
	resp := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

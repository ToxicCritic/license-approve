package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// In-memory stores для простоты
var (
	users              = map[string]*User{}
	authorizationCodes = map[string]*AuthorizationCode{}
	accessTokens       = map[string]*AccessToken{}
	refreshTokens      = map[string]*RefreshToken{}
)

// Инициализация с одним пользователем для демонстрации
func init() {
	users["user-1"] = &User{
		ID:       "user-1",
		Username: "admin",
		Password: "password",
	}
}

// Обрабатывает запросы на авторизацию
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on %s", r.Method, r.URL.Path)

	// Отображение формы авторизации
	if r.Method == "GET" {
		responseType := r.URL.Query().Get("response_type")
		clientID := r.URL.Query().Get("client_id")
		redirectURI := r.URL.Query().Get("redirect_uri")
		state := r.URL.Query().Get("state")
		scope := r.URL.Query().Get("scope")

		log.Printf("Authorization Request - Response Type: %s, Client ID: %s, Redirect URI: %s, State: %s, Scope: %s",
			responseType, clientID, redirectURI, state, scope)

		validID, _ := validateClient(clientID, "")
		if !validID {
			log.Printf("Invalid client ID: %s", clientID)
			http.Error(w, "Invalid client ID", http.StatusUnauthorized)
			return
		}
		// Проверка типа ответа
		if responseType != "code" {
			log.Printf("Unsupported response type: %s", responseType)
			http.Error(w, "Unsupported response type", http.StatusBadRequest)
			return
		}

		html := fmt.Sprintf(`
		<form method="POST" action="/authorize">
		<input type="hidden" name="response_type" value="%s" />
		<input type="hidden" name="client_id" value="%s" />
		<input type="hidden" name="redirect_uri" value="%s" />
		<input type="hidden" name="state" value="%s" />
		<input type="hidden" name="scope" value="%s" />
		<label>Username:</label><input type="text" name="username"/><br/>
		<label>Password:</label><input type="password" name="password"/><br/>
		<input type="submit" value="Authorize" />
		</form>
		`,
			responseType, clientID, redirectURI, state, scope)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		log.Println("Displayed authorization form")
		return
	}

	// Обработка POST запроса для авторизации
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		responseType := r.FormValue("response_type")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		state := r.FormValue("state")
		scope := r.FormValue("scope")

		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("Authorization Request - Response Type: %s, Client ID: %s, Redirect URI: %s, State: %s, Scope: %s",
			responseType, clientID, redirectURI, state, scope)

		if responseType != "code" {
			log.Printf("Unsupported response type: %s", responseType)
			http.Error(w, "Unsupported response type", http.StatusBadRequest)
			return
		}

		validID, _ := validateClient(clientID, "")
		if !validID {
			log.Printf("Invalid client ID: %s", clientID)
			http.Error(w, "Invalid client ID", http.StatusUnauthorized)
			return
		}

		// Аутентификация пользователя
		user, err := findUser(username, password)
		if err != nil {
			log.Printf("Invalid credentials for user: %s", username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Генерируем authorization code
		code, err := generateRandomString(32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		authorizationCodes[code] = &AuthorizationCode{
			Code:        code,
			ClientID:    clientID,
			UserID:      user.ID,
			RedirectURI: redirectURI,
			Expiry:      time.Now().Add(10 * time.Minute),
		}
		log.Printf("Generated authorization code: %s for user: %s", code, user.Username)

		// Редиректим обратно
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		log.Printf("Redirected to %s with code and state", redirectURL)

	}

	log.Printf("Method not allowed: %s", r.Method)
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// Обрабатывает запросы на получение токенов
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on %s", r.Method, r.URL.Path)

	// Проверка метода
	if r.Method != "POST" {
		log.Printf("Unsupported method for /token: %s", r.Method)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Разбор формы (grant_type, code, refresh_token и т.д.)
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form data: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	log.Printf("Token Request - Grant Type: %s", grantType)

	// Если grant_type не указан, но есть параметр token (т. е. "introspection")
	if grantType == "" && r.FormValue("token") != "" {
		handleIntrospection(w, r)
		return
	}

	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		handleRefreshTokenGrant(w, r)
	default:
		log.Printf("Unsupported grant type: %s", grantType)
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
	}
}

// Обрабатывает grant_type=authorization_code
func handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	log.Printf("AuthorizationCodeGrant - code=%s, redirect_uri=%s, client_id=%s",
		code, redirectURI, clientID)

	validID, validSecret := validateClient(clientID, clientSecret)
	if !validID {
		log.Printf("Invalid client ID: %s", clientID)
		http.Error(w, "Invalid client ID", http.StatusUnauthorized)
		return
	}
	if !validSecret {
		log.Printf("Invalid client Secret: %s", clientSecret)
		http.Error(w, "Invalid client Secret", http.StatusUnauthorized)
	}

	// Проверка authorization code
	authCode, exists := authorizationCodes[code]
	if !exists {
		log.Printf("Authorization code not found: %s", code)
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}
	if authCode.Expiry.Before(time.Now()) {
		log.Printf("Authorization code expired: %s", code)
		http.Error(w, "Authorization code expired", http.StatusBadRequest)
		return
	}
	if authCode.ClientID != clientID {
		log.Printf("Auth code client mismatch: code.ClientID=%s, req.ClientID=%s", authCode.ClientID, clientID)
		http.Error(w, "Invalid authorization code (client mismatch)", http.StatusBadRequest)
		return
	}
	if authCode.RedirectURI != redirectURI {
		log.Printf("Auth code redirect mismatch: code.RedirectURI=%s, req.RedirectURI=%s", authCode.RedirectURI, redirectURI)
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// Удаляем использованный authorization code (one-time use)
	delete(authorizationCodes, code)
	log.Printf("Authorization code %s used and deleted", code)

	// Генерация нового access_token и refresh_token
	accessToken, err := generateRandomString(32)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	refreshToken, err := generateRandomString(32)
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	at := &AccessToken{
		Token:        accessToken,
		UserID:       authCode.UserID,
		ClientID:     clientID,
		Expiry:       now.Add(15 * time.Second), // 15 секунд для теста рефреша
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
	}
	rt := &RefreshToken{
		Token:    refreshToken,
		UserID:   authCode.UserID,
		ClientID: clientID,
		Expiry:   now.Add(24 * time.Hour), // сутки
	}

	accessTokens[accessToken] = at
	refreshTokens[refreshToken] = rt

	log.Printf("Issued access token: %s and refresh token: %s for user ID: %s", accessToken, refreshToken, authCode.UserID)

	// Формируем JSON-ответ
	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    60, // 1 час
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("Token response sent for code=%s", code)
}

// Обрабатывает grant_type=refresh_token
func handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	log.Printf("RefreshTokenGrant - refresh_token=%s, client_id=%s", refreshToken, clientID)

	validID, validSecret := validateClient(clientID, clientSecret)
	if !validID {
		log.Printf("Invalid client ID: %s", clientID)
		http.Error(w, "Invalid client ID", http.StatusUnauthorized)
		return
	}
	if !validSecret {
		log.Printf("Invalid client Secret: %s", clientSecret)
		http.Error(w, "Invalid client Secret", http.StatusUnauthorized)
	}

	refToken, exists := refreshTokens[refreshToken]
	if !exists {
		log.Printf("Refresh token not found: %s", refreshToken)
		http.Error(w, "Invalid refresh token", http.StatusBadRequest)
		return
	}
	if refToken.Expiry.Before(time.Now()) {
		log.Printf("Refresh token expired: %s", refreshToken)
		http.Error(w, "Refresh token expired", http.StatusBadRequest)
		return
	}
	if refToken.ClientID != clientID {
		log.Printf("Refresh token client mismatch: token.ClientID=%s, req.ClientID=%s", refToken.ClientID, clientID)
		http.Error(w, "Invalid refresh token", http.StatusBadRequest)
		return
	}

	// Генерация новых access token и refresh token
	newAccessToken, err := generateRandomString(32)
	if err != nil {
		log.Printf("Error generating new access token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	newRefreshToken, err := generateRandomString(32)
	if err != nil {
		log.Printf("Error generating new refresh token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	at := &AccessToken{
		Token:        newAccessToken,
		UserID:       refToken.UserID,
		ClientID:     clientID,
		Expiry:       now.Add(15 * time.Second), // 15 секунд для теста рефреша
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
	}
	rt := &RefreshToken{
		Token:    newRefreshToken,
		UserID:   refToken.UserID,
		ClientID: clientID,
		Expiry:   now.Add(24 * time.Hour),
	}

	accessTokens[newAccessToken] = at
	refreshTokens[newRefreshToken] = rt

	// Удаляем старый refresh token
	delete(refreshTokens, refreshToken)
	log.Printf("Refresh token %s used and deleted", refreshToken)

	log.Printf("Issued new access token: %s and refresh token: %s for user ID: %s",
		newAccessToken, newRefreshToken, refToken.UserID)

	resp := map[string]interface{}{
		"access_token":  newAccessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": newRefreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("Token response sent for new access token: %s", newAccessToken)
}

// Обрабатывает запросы на получение информации о пользователе
func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request on %s", r.Method, r.URL.Path)

	// Извлечение Bearer токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("Missing Authorization header")
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	var accessToken string
	fmt.Sscanf(authHeader, "Bearer %s", &accessToken)

	log.Printf("UserInfo Request - Access Token: %s", accessToken)

	// Проверка access token
	token, exists := accessTokens[accessToken]
	if !exists {
		log.Printf("Invalid access token: %s", accessToken)
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}
	if token.Expiry.Before(time.Now()) {
		log.Printf("Access token expired: %s", accessToken)
		http.Error(w, "Access token expired", http.StatusUnauthorized)
		return
	}

	// Получение информации о пользователе
	user, exists := users[token.UserID]
	if !exists {
		log.Printf("User not found for access token: %s", accessToken)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Подготовка ответа
	resp := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("UserInfo response sent for user: %s", user.Username)
}

// Обрабатывает «интроспекцию» (проверку действия access_token),
// если grant_type не указан, но есть поле "token" (r.FormValue("token")).
func handleIntrospection(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	log.Printf("Introspection request for token=%s", token)

	at, exists := accessTokens[token]
	if !exists {
		// нет такого токена
		log.Printf("Introspection: no such access token: %s", token)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Active bool `json:"active"`
		}{Active: false})
		return
	}

	// Проверяем срок
	now := time.Now()
	if at.Expiry.After(now) {
		// Токен не истёк
		resp := struct {
			Active       bool   `json:"active"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token,omitempty"`
			ExpiresIn    int64  `json:"expires_in,omitempty"`
			TokenType    string `json:"token_type,omitempty"`
		}{
			Active:       true,
			AccessToken:  at.Token,
			RefreshToken: at.RefreshToken,
			ExpiresIn:    int64(at.Expiry.Sub(now).Seconds()),
			TokenType:    at.TokenType,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		log.Printf("Introspection: token is active, returning info for %s", token)
		return
		// } else {
		// 	// Токен истёк, пробуем "автоматический" рефреш
		// 	refresh := at.RefreshToken
		// 	log.Printf("Introspection: token expired. Attempt auto-refresh with refresh_token=%s", refresh)

		// 	// Находим refreshToken в refreshTokens
		// 	rt, exists := refreshTokens[refresh]
		// 	if !exists {
		// 		log.Printf("No such refresh token: %s => cannot auto-refresh", refresh)
		// 		w.Header().Set("Content-Type", "application/json")
		// 		json.NewEncoder(w).Encode(struct {
		// 			Active bool `json:"active"`
		// 		}{Active: false})
		// 		return
		// 	}
		// 	if rt.Expiry.Before(now) {
		// 		log.Printf("Refresh token expired: %s => cannot auto-refresh", refresh)
		// 		w.Header().Set("Content-Type", "application/json")
		// 		json.NewEncoder(w).Encode(struct {
		// 			Active bool `json:"active"`
		// 		}{Active: false})
		// 		return
		// 	}
		// 	// Удаляем старый tokens
		// 	delete(accessTokens, token)
		// 	delete(refreshTokens, refresh)

		// 	// Генерируем новые tokens
		// 	newAccessToken, err := generateRandomString(32)
		// 	if err != nil {
		// 		http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
		// 		return
		// 	}
		// 	newRefreshToken, err := generateRandomString(32)
		// 	if err != nil {
		// 		http.Error(w, "Failed to generate new refresh token", http.StatusInternalServerError)
		// 		return
		// 	}
		// 	at2 := &AccessToken{
		// 		Token:        newAccessToken,
		// 		UserID:       rt.UserID,
		// 		ClientID:     rt.ClientID,
		// 		Expiry:       now.Add(1 * time.Hour),
		// 		TokenType:    "Bearer",
		// 		RefreshToken: newRefreshToken,
		// 	}
		// 	rt2 := &RefreshToken{
		// 		Token:    newRefreshToken,
		// 		UserID:   rt.UserID,
		// 		ClientID: rt.ClientID,
		// 		Expiry:   now.Add(24 * time.Hour),
		// 	}
		// 	accessTokens[newAccessToken] = at2
		// 	refreshTokens[newRefreshToken] = rt2

		// 	log.Printf("Auto-refresh succeeded. newAccessToken=%s, newRefreshToken=%s", newAccessToken, newRefreshToken)

		// 	resp := struct {
		// 		Active       bool   `json:"active"`
		// 		AccessToken  string `json:"access_token"`
		// 		RefreshToken string `json:"refresh_token"`
		// 		ExpiresIn    int64  `json:"expires_in"`
		// 		TokenType    string `json:"token_type"`
		// 	}{
		// 		Active:       true,
		// 		AccessToken:  newAccessToken,
		// 		RefreshToken: newRefreshToken,
		// 		ExpiresIn:    int64(at2.Expiry.Sub(now).Seconds()),
		// 		TokenType:    "Bearer",
		// 	}
		// 	w.Header().Set("Content-Type", "application/json")
		// 	json.NewEncoder(w).Encode(resp)
		// 	log.Printf("Introspection: returning auto-refreshed token for old token=%s", token)
	}
}

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

		log.Printf("r.PostForm: %#v\n", r.PostForm)
		log.Printf("r.Form: %#v\n", r.Form)

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

	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form data: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	log.Printf("Token Request - Grant Type: %s", grantType)
	if grantType == "" && r.FormValue("token") != "" {
		// допустим, считаем что это introspection
		activeResp := struct {
			Active bool `json:"active"`
		}{
			Active: true, // или false, если хотим сказать "invalid"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(activeResp)
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

	log.Printf("Authorization Code Grant - Code: %s, Redirect URI: %s, Client ID: %s", code, redirectURI, clientID)

	// Проверка client credentials
	validID, _ := validateClient(clientID, clientSecret)
	if !validID {
		log.Printf("Invalid client ID: %s", clientID)
		http.Error(w, "Invalid client ID", http.StatusUnauthorized)
		return
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
		log.Printf("Authorization code client ID mismatch - Expected: %s, Got: %s", authCode.ClientID, clientID)
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}
	if authCode.RedirectURI != redirectURI {
		log.Printf("Authorization code redirect URI mismatch - Expected: %s, Got: %s", authCode.RedirectURI, redirectURI)
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	delete(authorizationCodes, code)
	log.Printf("Authorization code %s used and deleted", code)

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

	log.Printf("Issued access token: %s and refresh token: %s for user ID: %s", accessToken, refreshToken, authCode.UserID)

	// Подготовка ответа
	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Printf("Token response sent for access token: %s", accessToken)
}

// Обрабатывает grant_type=refresh_token
func handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	log.Printf("Refresh Token Grant - Refresh Token: %s, Client ID: %s", refreshToken, clientID)

	// Проверка client credentials
	validID, _ := validateClient(clientID, clientSecret)
	if !validID {
		log.Printf("Invalid client ID: %s", clientID)
		http.Error(w, "Invalid client ID", http.StatusUnauthorized)
		return
	}

	// Проверка refresh token
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
		log.Printf("Refresh token client ID mismatch - Expected: %s, Got: %s", refToken.ClientID, clientID)
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
	log.Printf("Refresh token %s used and deleted", refreshToken)

	log.Printf("Issued new access token: %s and refresh token: %s for user ID: %s", newAccessToken, newRefreshToken, refToken.UserID)

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

// mock-oauth-server/internal/handlers/handlers.go

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"mock-oauth-server/internal/models"
)

// Config for passing client credentials from main
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// local variable to store config
var localConfig Config

func SetConfig(cfg Config) {
	localConfig = cfg
}

// ============ /authorize Handler ============
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleAuthorizeGet(w, r)
	} else if r.Method == http.MethodPost {
		handleAuthorizePost(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAuthorizeGet(w http.ResponseWriter, r *http.Request) {
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")

	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	// (Optional) check clientID == localConfig.ClientID ?

	htmlForm := fmt.Sprintf(`
<html>
<body>
<h2>Mock OAuth2 Authorization</h2>
<form method="POST" action="/authorize">
  <input type="hidden" name="response_type" value="%s" />
  <input type="hidden" name="client_id" value="%s" />
  <input type="hidden" name="redirect_uri" value="%s" />
  <input type="hidden" name="state" value="%s" />
  <input type="hidden" name="scope" value="%s" />
  <label>Username:</label><input type="text" name="username"/><br/>
  <label>Password:</label><input type="password" name="password"/><br/>
  <button type="submit">Authorize</button>
</form>
</body>
</html>
`, responseType, clientID, redirectURI, state, scope)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlForm))
}

func handleAuthorizePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	responseType := r.FormValue("response_type")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	scope := r.FormValue("scope")

	username := r.FormValue("username")
	password := r.FormValue("password")

	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	// check user
	models.InMemoryMutex.Lock()
	user, ok := models.Users[username]
	models.InMemoryMutex.Unlock()
	if !ok || user.Password != password {
		http.Error(w, "invalid user credentials", http.StatusUnauthorized)
		return
	}

	codeVal, _ := models.GenerateRandomString(32)
	now := time.Now()
	authCode := &models.AuthorizationCode{
		Code:        codeVal,
		ClientID:    clientID,
		UserID:      username,
		RedirectURI: redirectURI,
		Scope:       scope,
		Expiry:      now.Add(5 * time.Minute),
	}

	models.InMemoryMutex.Lock()
	models.AuthCodes[codeVal] = authCode
	models.InMemoryMutex.Unlock()

	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, codeVal, state)
	log.Printf("[MOCK-OAUTH] Issued code=%s to user=%s scope=%s -> redirect=%s", codeVal, username, scope, redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// ============ /token Handler ============
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	grantType := r.FormValue("grant_type")
	log.Printf("[TOKEN] grant_type=%s\n", grantType)

	switch grantType {
	case "authorization_code":
		handleAuthorizationCodeGrant(w, r)
	case "password":
		handlePasswordGrant(w, r)
	case "refresh_token":
		handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant ...
func handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	codeVal := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// optional: check client creds
	if clientID != localConfig.ClientID || clientSecret != localConfig.ClientSecret {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	models.InMemoryMutex.Lock()
	authCode, ok := models.AuthCodes[codeVal]
	if !ok {
		models.InMemoryMutex.Unlock()
		http.Error(w, "invalid authorization code", http.StatusBadRequest)
		return
	}
	if time.Now().After(authCode.Expiry) {
		delete(models.AuthCodes, codeVal)
		models.InMemoryMutex.Unlock()
		http.Error(w, "authorization code expired", http.StatusBadRequest)
		return
	}
	if authCode.ClientID != clientID {
		models.InMemoryMutex.Unlock()
		http.Error(w, "client_id mismatch", http.StatusBadRequest)
		return
	}
	if authCode.RedirectURI != redirectURI {
		models.InMemoryMutex.Unlock()
		http.Error(w, "redirect_uri mismatch", http.StatusBadRequest)
		return
	}
	// use once
	delete(models.AuthCodes, codeVal)

	accessT, _ := models.GenerateRandomString(32)
	refreshT, _ := models.GenerateRandomString(32)
	now := time.Now()
	at := &models.AccessToken{
		Token:        accessT,
		UserID:       authCode.UserID,
		ClientID:     clientID,
		TokenType:    "Bearer",
		Expiry:       now.Add(1 * time.Hour),
		RefreshToken: refreshT,
	}
	rt := &models.RefreshToken{
		Token:    refreshT,
		UserID:   authCode.UserID,
		ClientID: clientID,
		Expiry:   now.Add(24 * time.Hour),
	}
	models.AccessTokens[accessT] = at
	models.RefreshTokens[refreshT] = rt
	models.InMemoryMutex.Unlock()

	resp := map[string]interface{}{
		"access_token":  accessT,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshT,
		"scope":         authCode.Scope,
	}
	w.Header().Set("Content-Type", "application/json")
	log.Printf("[TOKEN-CODE] user=%s -> access_token=%s refresh_token=%s", authCode.UserID, accessT, refreshT)
	writeJSON(w, resp)
}

// handlePasswordGrant ...
func handlePasswordGrant(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	scope := r.FormValue("scope")

	if clientID != localConfig.ClientID || clientSecret != localConfig.ClientSecret {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	models.InMemoryMutex.Lock()
	user, ok := models.Users[username]
	models.InMemoryMutex.Unlock()
	if !ok || user.Password != password {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	accessT, _ := models.GenerateRandomString(32)
	refreshT, _ := models.GenerateRandomString(32)
	now := time.Now()
	at := &models.AccessToken{
		Token:        accessT,
		UserID:       username,
		ClientID:     clientID,
		TokenType:    "Bearer",
		Expiry:       now.Add(1 * time.Hour),
		RefreshToken: refreshT,
	}
	rt := &models.RefreshToken{
		Token:    refreshT,
		UserID:   username,
		ClientID: clientID,
		Expiry:   now.Add(24 * time.Hour),
	}

	models.InMemoryMutex.Lock()
	models.AccessTokens[accessT] = at
	models.RefreshTokens[refreshT] = rt
	models.InMemoryMutex.Unlock()

	resp := map[string]interface{}{
		"access_token":  accessT,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshT,
		"scope":         scope,
	}
	log.Printf("[TOKEN-PASSWORD] user=%s -> access_token=%s refresh_token=%s", username, accessT, refreshT)
	writeJSON(w, resp)
}

// handleRefreshTokenGrant ...
func handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshVal := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	if clientID != localConfig.ClientID || clientSecret != localConfig.ClientSecret {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	models.InMemoryMutex.Lock()
	defer models.InMemoryMutex.Unlock()

	rt, ok := models.RefreshTokens[refreshVal]
	if !ok {
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
		return
	}
	if time.Now().After(rt.Expiry) {
		delete(models.RefreshTokens, refreshVal)
		http.Error(w, "refresh token expired", http.StatusBadRequest)
		return
	}

	// generate new tokens
	newAccessT, _ := models.GenerateRandomString(32)
	newRefreshT, _ := models.GenerateRandomString(32)
	now := time.Now()
	at := &models.AccessToken{
		Token:        newAccessT,
		UserID:       rt.UserID,
		ClientID:     rt.ClientID,
		TokenType:    "Bearer",
		Expiry:       now.Add(1 * time.Hour),
		RefreshToken: newRefreshT,
	}
	rt2 := &models.RefreshToken{
		Token:    newRefreshT,
		UserID:   rt.UserID,
		ClientID: rt.ClientID,
		Expiry:   now.Add(24 * time.Hour),
	}
	// remove old
	delete(models.RefreshTokens, refreshVal)
	models.AccessTokens[newAccessT] = at
	models.RefreshTokens[newRefreshT] = rt2

	resp := map[string]interface{}{
		"access_token":  newAccessT,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": newRefreshT,
	}
	log.Printf("[TOKEN-REFRESH] user=%s oldRefresh=%s => newAccess=%s newRefresh=%s", rt.UserID, refreshVal, newAccessT, newRefreshT)
	writeJSON(w, resp)
}

// ============ /userinfo Handler ============
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[MOCK-OAUTH] /userinfo")

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	models.InMemoryMutex.Lock()
	at, ok := models.AccessTokens[tokenStr]
	if !ok {
		models.InMemoryMutex.Unlock()
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	if time.Now().After(at.Expiry) {
		delete(models.AccessTokens, tokenStr)
		models.InMemoryMutex.Unlock()
		http.Error(w, "access token expired", http.StatusUnauthorized)
		return
	}
	userID := at.UserID
	models.InMemoryMutex.Unlock()

	resp := map[string]string{
		"id":       userID,
		"username": userID,
	}
	writeJSON(w, resp)
	log.Printf("[USERINFO] Provided user info for %s", userID)
}

// Helper to write JSON
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

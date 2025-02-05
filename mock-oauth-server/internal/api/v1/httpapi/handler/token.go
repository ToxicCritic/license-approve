package handler

import (
	"encoding/json"
	"mime"
	"mock-oauth-server/internal/repository/inmem"
	"net/http"
	"time"
)

// TokenHandler обрабатывает /oauth/token.
// Сюда входят:
//   - grant_type=authorization_code
//   - grant_type=refresh_token
//   - интроспекция (если grant_type пуст, но есть token=?)
func (h *Handler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	ct := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if mediaType != "application/x-www-form-urlencoded" && mediaType != "multipart/form-data" {
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}
	grantType := r.FormValue("grant_type")
	tokenParam := r.FormValue("token")

	if grantType == "" && tokenParam != "" {
		h.introspectToken(w, r)
		return
	}

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (h *Handler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	codeVal := r.FormValue("code")
	clientID := r.FormValue("client_id")
	// clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	ac, ok := h.store.AuthorizationCodes[codeVal]
	if !ok {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	if ac.ClientID != clientID {
		http.Error(w, "client mismatch", http.StatusBadRequest)
		return
	}
	if ac.RedirectURI != redirectURI {
		http.Error(w, "redirect mismatch", http.StatusBadRequest)
		return
	}
	if time.Now().After(ac.Expiry) {
		http.Error(w, "code expired", http.StatusBadRequest)
		return
	}
	// Удаляем использованный code
	delete(h.store.AuthorizationCodes, codeVal)

	// Генерим access_token и refresh_token
	accessToken := h.store.GenerateRandomString(32)
	refreshToken := h.store.GenerateRandomString(32)

	h.store.AccessTokens[accessToken] = &inmem.AccessToken{
		Token:     accessToken,
		UserID:    ac.UserID,
		ClientID:  clientID,
		Expiry:    time.Now().Add(1 * time.Minute),
		TokenType: "Bearer",
	}
	h.store.RefreshTokens[refreshToken] = &inmem.RefreshToken{
		Token:    refreshToken,
		UserID:   ac.UserID,
		ClientID: clientID,
		Expiry:   time.Now().Add(24 * time.Hour),
	}

	resp := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    60,
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshVal := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")

	rt, ok := h.store.RefreshTokens[refreshVal]
	if !ok {
		http.Error(w, "invalid refresh_token", http.StatusBadRequest)
		return
	}
	if time.Now().After(rt.Expiry) {
		http.Error(w, "refresh_token expired", http.StatusBadRequest)
		return
	}
	if rt.ClientID != clientID {
		http.Error(w, "client mismatch in refresh_token", http.StatusBadRequest)
		return
	}

	newAccess := h.store.GenerateRandomString(32)
	newRefresh := h.store.GenerateRandomString(32)

	h.store.AccessTokens[newAccess] = &inmem.AccessToken{
		Token:     newAccess,
		UserID:    rt.UserID,
		ClientID:  clientID,
		Expiry:    time.Now().Add(1 * time.Minute),
		TokenType: "Bearer",
	}
	h.store.RefreshTokens[newRefresh] = &inmem.RefreshToken{
		Token:    newRefresh,
		UserID:   rt.UserID,
		ClientID: clientID,
		Expiry:   time.Now().Add(24 * time.Hour),
	}
	// старый refresh_token удаляем
	delete(h.store.RefreshTokens, refreshVal)

	resp := map[string]interface{}{
		"access_token":  newAccess,
		"token_type":    "Bearer",
		"expires_in":    60,
		"refresh_token": newRefresh,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) introspectToken(w http.ResponseWriter, r *http.Request) {
	tokVal := r.FormValue("token")
	at, ok := h.store.AccessTokens[tokVal]
	active := false
	if ok && time.Now().Before(at.Expiry) {
		active = true
	}
	resp := map[string]bool{"active": active}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

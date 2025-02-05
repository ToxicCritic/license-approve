// internal/api/v1/httpapi/handler/token.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// handle /oauth/token
func (h *Handler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	grantType := r.FormValue("grant_type")
	switch grantType {
	case "authorization_code":
		// (mock) сразу выдаём token
		h.handleAuthCode(w, r)
	case "refresh_token":
		h.handleRefreshToken(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (h *Handler) handleAuthCode(w http.ResponseWriter, r *http.Request) {
	// Для упрощения "code" не проверяем, просто выдаём токен
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	cli, ok := h.store.FindClient(clientID, clientSecret)
	if !ok {
		http.Error(w, "invalid client creds", http.StatusUnauthorized)
		return
	}
	// userID = "u1"
	at, rt, err := h.jwtSvc.GenerateTokenPair("u1", cli.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"access_token":  at,
		"refresh_token": rt,
		"token_type":    "Bearer",
		"expires_in":    3600,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	rt := r.FormValue("refresh_token")
	at, rt2, err := h.jwtSvc.RefreshToken(rt)
	if err != nil {
		http.Error(w, fmt.Sprintf("refresh error: %v", err), http.StatusBadRequest)
		return
	}
	resp := map[string]interface{}{
		"access_token":  at,
		"refresh_token": rt2,
		"token_type":    "Bearer",
		"expires_in":    3600,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

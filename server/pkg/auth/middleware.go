package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	"github.com/spf13/viper"
)

// AuthMiddleware защищает маршруты, требующие аутентификации и авторизации.
func AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("[AuthMiddleware] Intercepting request:", r.Method, r.URL.Path)

			// 1. Получаем сессию
			session, err := Store.Get(r, "auth-session")
			if err != nil {
				log.Printf("[AuthMiddleware] Store.Get error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// 2. Проверяем, аутентифицирован ли пользователь
			auth, ok := session.Values["authenticated"].(bool)
			if !ok || !auth {
				log.Println("[AuthMiddleware] Not authenticated: session has no 'authenticated' flag")
				http.Error(w, "Forbidden: Not authenticated", http.StatusForbidden)
				return
			}
			log.Println("[AuthMiddleware] 'authenticated' = true")

			// 3. Достаём userSession (AccessToken etc.) из session
			userSession, ok := session.Values["user"].(*UserSession)
			if !ok || userSession.AccessToken == "" {
				log.Println("[AuthMiddleware] No user session or empty AccessToken in session")
				http.Error(w, "Forbidden: Token not found", http.StatusForbidden)
				return
			}
			log.Printf("[AuthMiddleware] Found user session with AccessToken=%q Expiry=%v RefreshToken=%q\n",
				userSession.AccessToken, userSession.Expiry, userSession.RefreshToken)

			now := time.Now()
			// 4. Если токен истёк, делаем refresh
			if userSession.Expiry.Before(now) {
				log.Printf("[AuthMiddleware] Access token is expired at %v (now=%v). Attempt refresh...",
					userSession.Expiry, now)

				newToken, err := refreshAccessToken(userSession.RefreshToken)
				if err != nil {
					log.Printf("[AuthMiddleware] refreshAccessToken failed: %v", err)
					http.Error(w, "Forbidden: Token expired and refresh failed", http.StatusForbidden)
					return
				}

				userSession.AccessToken = newToken.AccessToken
				userSession.RefreshToken = newToken.RefreshToken
				userSession.TokenType = newToken.TokenType
				userSession.Expiry = newToken.Expiry

				session.Values["user"] = userSession
				if err := session.Save(r, w); err != nil {
					log.Printf("[AuthMiddleware] session.Save failed: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				log.Printf("[AuthMiddleware] Token was refreshed successfully. New expiry=%v\n", userSession.Expiry)
			} else {
				log.Printf("[AuthMiddleware] Access token still valid. Expires at %v (now=%v)\n",
					userSession.Expiry, now)
			}

			// 5. Проверяем токен на валидность (introspection)
			valid, err := validateAccessToken(userSession.AccessToken)
			if err != nil {
				log.Printf("[AuthMiddleware] validateAccessToken returned error: %v", err)
				http.Error(w, "Forbidden: Invalid token (error)", http.StatusForbidden)
				return
			}
			if !valid {
				log.Println("[AuthMiddleware] validateAccessToken says token is not active/valid")
				http.Error(w, "Forbidden: Invalid token", http.StatusForbidden)
				return
			}

			log.Println("[AuthMiddleware] Token is valid. Proceed...")
			next.ServeHTTP(w, r)
		})
	}
}

// validateAccessToken отправляет запрос на introspection endpoint, проверяя, что accessToken активен.
func validateAccessToken(accessToken string) (bool, error) {
	log.Printf("[validateAccessToken] Checking token=%q\n", accessToken)

	tokenURL := viper.GetString("OAUTH_TOKEN_URL")
	if tokenURL == "" {
		return false, fmt.Errorf("[validateAccessToken] OAUTH_TOKEN_URL not set in viper config")
	}

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return false, fmt.Errorf("[validateAccessToken] failed to create token introspection request: %v", err)
	}

	log.Printf("[validateAccessToken] Using OAuthConfig.ClientID=%q OAuthConfig.ClientSecret=%q\n",
		OAuthConfig.ClientID, OAuthConfig.ClientSecret)
	req.SetBasicAuth(OAuthConfig.ClientID, OAuthConfig.ClientSecret)

	q := req.URL.Query()
	q.Add("token", accessToken)
	q.Add("token_type_hint", "access_token")
	req.URL.RawQuery = q.Encode()

	log.Printf("[validateAccessToken] POST -> %s?%s\n", tokenURL, req.URL.RawQuery)

	// Если самоподписанный сертификат у authserver:
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: insecureTransport, Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("[validateAccessToken] failed to introspect token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("[validateAccessToken] introspection returned status: %s", resp.Status)
	}

	var introspectResp struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&introspectResp); err != nil {
		return false, fmt.Errorf("[validateAccessToken] failed to decode introspection JSON: %v", err)
	}

	log.Printf("[validateAccessToken] Introspection response: Active=%v\n", introspectResp.Active)
	return introspectResp.Active, nil
}

// refreshAccessToken обновляет токен доступа с помощью refreshToken (через oauth2.TokenSource).
func refreshAccessToken(refreshToken string) (*oauth2.Token, error) {
	log.Printf("[refreshAccessToken] Attempting refresh with refresh_token=%q\n", refreshToken)

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureClient := &http.Client{Transport: insecureTransport}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, insecureClient)

	newToken, err := OAuthConfig.TokenSource(ctx, token).Token()
	if err != nil {
		log.Printf("[refreshAccessToken] TokenSource().Token() error: %v\n", err)
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	log.Printf("[refreshAccessToken] Successfully refreshed. New AccessToken=%q Expiry=%v\n",
		newToken.AccessToken, newToken.Expiry)
	return newToken, nil
}

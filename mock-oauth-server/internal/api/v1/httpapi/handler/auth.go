package handler

import (
	"fmt"
	"log"
	"mock-oauth-server/internal/repository/inmem"
	"net/http"
	"net/url"
	"time"
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

	// Примитивная форма
	html := fmt.Sprintf(`
<html>
 <body>
   <h3>Mock /authorize</h3>
   <form method="POST">
     <input type="hidden" name="response_type" value="%s">
     <input type="hidden" name="client_id" value="%s">
     <input type="hidden" name="redirect_uri" value="%s">
     <input type="hidden" name="state" value="%s">
     <input type="hidden" name="scope" value="%s">
     <label>Username:</label><input type="text" name="username"><br/>
     <label>Password:</label><input type="password" name="password"><br/>
     <input type="submit" value="Authorize">
   </form>
 </body>
</html>
`, responseType, clientID, redirectURI, state, scope)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *Handler) handleAuthorizePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}
	responseType := r.FormValue("response_type")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != "admin" || password != "password" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}

	// Генерируем code
	code := h.store.GenerateRandomString(32)
	h.store.AuthorizationCodes[code] = &inmem.AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		UserID:      "user-1",
		RedirectURI: redirectURI,
		Expiry:      time.Now().Add(5 * time.Minute),
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirectURI", http.StatusBadRequest)
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

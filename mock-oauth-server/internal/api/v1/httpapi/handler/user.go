package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// пример, возвращаем список
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	usersList := h.usecases.GetAllUsers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usersList)
}

// пример создания
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if body.Username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}
	if err := h.usecases.CreateUser(body.Username, body.Password); err != nil {
		http.Error(w, "Cannot create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, `{"status":"ok"}`)
}

// заглушка
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userLogin := chi.URLParam(r, "UserLogin")
	http.Error(w, "UpdateUser not implemented (mock). userLogin="+userLogin, http.StatusNotImplemented)
}

// заглушка
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userLogin := chi.URLParam(r, "UserLogin")
	http.Error(w, "DeleteUser not implemented (mock). userLogin="+userLogin, http.StatusNotImplemented)
}

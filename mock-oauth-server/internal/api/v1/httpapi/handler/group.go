package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// пример
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	list := h.usecases.GetAllGroups()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// пример
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "Group name is empty", http.StatusBadRequest)
		return
	}
	err := h.usecases.CreateGroup(body.Name)
	if err != nil {
		http.Error(w, "Cannot create group: "+err.Error(), http.StatusConflict)
		return
	}
	fmt.Fprintln(w, `{"status":"ok"}`)
}

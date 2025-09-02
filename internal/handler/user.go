package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
)

// UserHandler handles user-related endpoints.
type UserHandler struct {
	db *database.Queries
}

// NewUserHandler creates a new user handler.
func NewUserHandler(db *database.Queries) *UserHandler {
	return &UserHandler{db: db}
}

// SearchUsers handles the user search endpoint.
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	users, err := h.db.SearchUsers(r.Context(), "%"+query+"%")
	if err != nil {
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetAllUsers handles getting all users.
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.db.GetAllUsers(r.Context())
    if err != nil {
        http.Error(w, "Failed to get all users", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

// GetUserByID handles getting a user by their ID.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    userIDParam := chi.URLParam(r, "id")
    userID, err := uuid.Parse(userIDParam)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    user, err := h.db.GetUserByID(r.Context(), userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
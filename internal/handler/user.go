package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mxhdiqaim/go-chat-app/internal/database"
)

// UserHandler handles user-related API requests.
type UserHandler struct {
    db *database.Queries
}

// NewUserHandler creates a new user handler.
func NewUserHandler(db *database.Queries) *UserHandler {
    return &UserHandler{db: db}
}

// SearchUsers handles searching for users by username.
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
        return
    }

    users, err := h.db.SearchUsers(r.Context(), "%" + query + "%")
    if err != nil {
        http.Error(w, "Failed to search for users", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}
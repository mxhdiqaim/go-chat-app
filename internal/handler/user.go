package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

// UserHandler handles user-related endpoints.
type UserHandler struct {
    db *database.Queries
}

// NewUserHandler creates a new user handler.
func NewUserHandler(db *database.Queries) *UserHandler {
    return &UserHandler{db: db}
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

// UpdateUser handles updating a user's account.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    // Get the authenticated user's ID from the JWT middleware
    authUserID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get the user ID from the URL
    userIDParam := chi.URLParam(r, "id")

    // Check if the authenticated user's ID matches the requested user ID
    if authUserID != userIDParam {
        http.Error(w, "Forbidden: You can only update your own account", http.StatusForbidden)
        return
    }

    // Parse the user ID from the URL
    userID, err := uuid.Parse(userIDParam)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Decode the request body to get new data
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Hash the new password before storing it
    hashedPassword, err := service.HashPassword(req.Password)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }

    updatedUser, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
        ID:       userID,
        Username: req.Username,
        Password: hashedPassword,
    })
    if err != nil {
        log.Println("Failed to update user:", err)
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedUser)
}

// DeleteUser handles deleting a user's account.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    // Get the authenticated user's ID from the JWT middleware
    authUserID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get the user ID from the URL
    userIDParam := chi.URLParam(r, "id")

    // Check if the authenticated user's ID matches the requested user ID
    if authUserID != userIDParam {
        http.Error(w, "Forbidden: You can only delete your own account", http.StatusForbidden)
        return
    }

    // Parse the user ID from the URL
    userID, err := uuid.Parse(userIDParam)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    if err := h.db.DeleteUser(r.Context(), userID); err != nil {
        log.Println("Failed to delete user:", err)
        http.Error(w, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
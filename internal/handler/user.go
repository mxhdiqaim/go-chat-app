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

// UpdateUserRequest defines the request body for updating a user.
type UpdateUserRequest struct {
    Username string `json:"username" example:"updateduser"`
    Password string `json:"password" example:"newpassword123"`
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users in the system.
// @Tags         users
// @Produce      json
// @Success      200  {array}   UserResponse
// @Failure      500  {string}  string "Failed to get users"
// @Router       /users [get]
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.db.GetAllUsers(r.Context())
    if err != nil {
        http.Error(w, "Failed to get all users", http.StatusInternalServerError)
        return
    }

    // Convert database models to response DTOs
    var responses []UserResponse

    for _, user := range users {
        responses = append(responses, UserResponse{
            ID:        user.ID,
            Username:  user.Username,
            CreatedAt: user.CreatedAt.Time,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(responses)
}

// GetUserByID godoc
// @Summary      Get a single user by ID
// @Description  Retrieves details for a specific user.
// @Tags         users
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200 {object}  UserResponse
// @Failure      400 {string}  string "Invalid user ID"
// @Failure      404 {string}  string "User not found"
// @Router       /users/{id} [get]
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

    response := UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        CreatedAt: user.CreatedAt.Time,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// SearchUsers godoc
// @Summary      Search for users
// @Description  Searches for users by username.
// @Tags         users
// @Produce      json
// @Param        q   query     string  true  "Search query"
// @Success      200 {array}   UserResponse
// @Failure      400 {string}  string "Query parameter 'q' is required"
// @Failure      500 {string}  string "Failed to search users"
// @Router       /users/search [get]
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

    var responses []UserResponse
    for _, user := range users {
        responses = append(responses, UserResponse{
            ID:        user.ID,
            Username:  user.Username,
            CreatedAt: user.CreatedAt.Time,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(responses)
}

// UpdateUser godoc
// @Summary      Update a user's account
// @Description  Updates a user's username and/or password. Users can only update their own account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "User ID"
// @Param        user  body      UpdateUserRequest  true  "User data to update"
// @Success      200   {object}  UserResponse
// @Failure      400   {string}  string "Invalid user ID or request body"
// @Failure      401   {string}  string "User not authenticated"
// @Failure      403   {string}  string "Forbidden: You can only update your own account"
// @Failure      500   {string}  string "Failed to update user"
// @Security     ApiKeyAuth
// @Router       /users/{id} [put]
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

    var req UpdateUserRequest
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

    response := UserResponse{
        ID:        updatedUser.ID,
        Username:  updatedUser.Username,
        CreatedAt: updatedUser.CreatedAt.Time,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// DeleteUser godoc
// @Summary      Delete a user's account
// @Description  Deletes a user's account. Users can only delete their own account.
// @Tags         users
// @Param        id  path      string  true  "User ID"
// @Success      204 {string}  string  "No Content"
// @Failure      400 {string}  string  "Invalid user ID"
// @Failure      401 {string}  string  "User not authenticated"
// @Failure      403 {string}  string  "Forbidden: You can only delete your own account"
// @Failure      500 {string}  string  "Failed to delete user"
// @Security     ApiKeyAuth
// @Router       /users/{id} [delete]
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
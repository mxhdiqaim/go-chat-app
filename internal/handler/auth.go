package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
    userService *service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService) *AuthHandler {
    return &AuthHandler{userService: userService}
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Use the new service function to hash the password
    hashedPassword, err := service.HashPassword(req.Password)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }

    user, err := h.userService.CreateUser(r.Context(), req.Username, hashedPassword)
    if err != nil {
        http.Error(w, "Registration failed", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// LoginUser handles user login and issues a JWT
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    user, err := h.userService.GetUserByUsername(r.Context(), req.Username)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

   // Use the new service function to check the password
    if !service.CheckPasswordHash(req.Password, user.Password) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    token, err := middleware.GenerateJWT(user.ID.String(), 24*time.Hour)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}
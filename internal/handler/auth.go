package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
    userService *service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService) *AuthHandler {
    return &AuthHandler{userService: userService}
}

// RegisterRequest defines the shape of the registration request body.
type RegisterRequest struct {
    Username string `json:"username" example:"newuser"`
    Password string `json:"password" example:"password123"`
}

// LoginRequest defines the shape of the login request body.
type LoginRequest struct {
    Username string `json:"username" example:"newuser"`
    Password string `json:"password" example:"password123"`
}

// UserResponse is the DTO for a user that is safe to send to clients.
type UserResponse struct {
    ID        uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
    Username  string    `json:"username" example:"newuser"`
    CreatedAt time.Time `json:"created_at" example:"2025-09-03T12:00:00Z"`
}

// LoginResponse defines the shape of the successful login response.
type LoginResponse struct {
    Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Create a new user with a username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      RegisterRequest  true  "User Registration Info"
// @Success      201   {object}  UserResponse
// @Failure      400   {string}  string "Invalid request body"
// @Failure      500   {string}  string "Registration failed"
// @Router       /register [post]
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest

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

    response := UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        CreatedAt: user.CreatedAt.Time,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

// LoginUser godoc
// @Summary      Log in a user
// @Description  Log in with username and password to receive a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest     true  "User Credentials"
// @Success      200          {object}  LoginResponse
// @Failure      400          {string}  string "Invalid request body"
// @Failure      401          {string}  string "Invalid credentials"
// @Failure      500          {string}  string "Failed to generate token"
// @Router       /login [post]
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
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
    json.NewEncoder(w).Encode(LoginResponse{Token: token})
}
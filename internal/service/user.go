package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides user-related business logic.
type UserService struct {
    db *database.Queries
}

// NewUserService creates a new UserService.
func NewUserService(db *database.Queries) *UserService {
    return &UserService{db: db}
}

// HashPassword hashes a user's password using bcrypt.
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a hashed password.
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// CreateUser creates a new user in the database.
func (s *UserService) CreateUser(ctx context.Context, username, hashedPassword string) (database.User, error) {
    return s.db.CreateUser(ctx, database.CreateUserParams{
        ID:       uuid.New(),
        Username: username,
        Password: hashedPassword,
    })
}

// GetUserByUsername retrieves a user by their username.
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (database.User, error) {
    return s.db.GetUserByUsername(ctx, username)
}

// GetUserByID retrieves a user by their ID.
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
    return s.db.GetUserByID(ctx, id)
}
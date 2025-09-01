package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides user-related business logic
type UserService struct {
    db *database.Queries
}

// NewUserService creates a new user service
func NewUserService(db *database.Queries) *UserService {
    return &UserService{db: db}
}

// CreateUser hashes the password and creates a new user
func (s *UserService) CreateUser(ctx context.Context, username, password string) (database.User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return database.User{}, fmt.Errorf("failed to hash password: %w", err)
    }

    id := uuid.New()
    
    user, err := s.db.CreateUser(ctx, database.CreateUserParams{
        ID:             id,
        Username:       username,
        PasswordHash:   string(hashedPassword),
    })
    if err != nil {
        return database.User{}, fmt.Errorf("failed to create user: %w", err)
    }
    return user, nil
}

// AuthenticateUser retrieves the user and compares the password
func (s *UserService) AuthenticateUser(ctx context.Context, username, password string) (database.User, error) {
    user, err := s.db.GetUserByUsername(ctx, username)
    if err != nil {
        return database.User{}, fmt.Errorf("user not found: %w", err)
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    if err != nil {
        return database.User{}, errors.New("invalid credentials")
    }

    return user, nil
}
package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string
const ContextUserIDKey contextKey = "userID"

// AuthMiddleware checks for a valid JWT and adds the user ID to the context
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the token from the Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header is required", http.StatusUnauthorized)
            return
        }

        // Split the header to get the token part
        headerParts := strings.Split(authHeader, " ")
        if len(headerParts) != 2 || headerParts[0] != "Bearer" {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }
        tokenString := headerParts[1]

        // Parse and validate the token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, http.ErrNotSupported
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Extract the user ID from the claims
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            userID, ok := claims["sub"].(string)
            if !ok {
                http.Error(w, "Invalid token claims", http.StatusUnauthorized)
                return
            }

            // Add the user ID to the request context
            ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
            r = r.WithContext(ctx)
            next.ServeHTTP(w, r)
        } else {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
    })
}
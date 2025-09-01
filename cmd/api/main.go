package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/handler"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

func main() {
    // Database Connection Pool
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgresql://postgres:password@localhost:5432/postgres"
    }
    dbPool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("Unable to create connection pool: %v\n", err)
    }
    defer dbPool.Close()
    dbQueries := database.New(dbPool)

    // Initialize Services and Handlers
    userService := service.NewUserService(dbQueries)
    authHandler := handler.NewAuthHandler(userService)

    // Router Setup
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Public Routes
    r.Post("/register", authHandler.RegisterUser)
    r.Post("/login", authHandler.LoginUser)

    // NOTE: For now, we will add the WebSocket route here, but it will be moved later
    //r.Get("/ws/{roomID}", ...)

    // Start Server
    port := ":8080"
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(port, r))
}
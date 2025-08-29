package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	// Import your generated SQLC code
	"github.com/mxhdiqaim/go-chat-app/database"
)

// Hub manages active clients and broadcasts messages.
// This is where you will discuss memory management.
type Hub struct {
    mu      sync.RWMutex // Protects concurrent access to clients map
    // Using a map of pointers to connections is memory-efficient.
    // The GC will clean up the client when the pointer is removed.
    clients map[*websocket.Conn]bool
}

func newHub() *Hub {
    return &Hub{
        clients: make(map[*websocket.Conn]bool),
    }
}

var hub = newHub()

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
    // --- DB Connection ---
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

    // --- Router with Chi ---
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // API Handlers (using sqlc generated functions)
    r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
        // Assume you get username from request body
        username := "testuser"
        user, err := dbQueries.CreateUser(r.Context(), database.CreateUserParams{
            ID:       uuid.New(),
            Username: username,
        })
        if err != nil {
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "User created: %s", user.Username)
    })

    // WebSocket Handler (using go routines and channels for concurrency)
    r.Get("/ws/{roomID}", func(w http.ResponseWriter, r *http.Request) {
        // You will use chi.URLParam here to get the roomID
        roomID := chi.URLParam(r, "roomID")
        log.Printf("Client is connecting to room %s", roomID)

        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("Failed to upgrade to WebSocket: %v", err)
            return
        }
        
        hub.mu.Lock()
        // Here we store the POINTER to the connection, not the value.
        // This is a memory management best practice.
        hub.clients[conn] = true
        hub.mu.Unlock()

        // This goroutine handles messages from the client concurrently.
        go func() {
            defer func() {
                hub.mu.Lock()
                delete(hub.clients, conn)
                hub.mu.Unlock()
                conn.Close()
            }()
            
            for {
                // Read messages from the client
                _, _, err := conn.ReadMessage()
                if err != nil {
                    log.Printf("Client disconnected: %v", err)
                    return
                }
                // Your broadcasting logic goes here
            }
        }()
    })

    // --- Server Start ---
    port := ":8080"
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(port, r))
}
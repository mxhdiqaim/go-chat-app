package main

import (
	"context"
	"encoding/json"

	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxhdiqaim/go-chat-app/database"
)

// Message is the JSON structure for a chat message
type Message struct {
	Body string `json:"body"`
}

// Hub manages active clients and broadcasts messages.
type Hub struct {
	mu        sync.RWMutex
	clients   map[*websocket.Conn]bool
	broadcast chan []byte // New: Channel to receive messages to broadcast
}

func newHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

var hub = newHub()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
    // DB Connection
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

	// Start the goroutine that listens for messages to broadcast
	go run()

    // Router with Chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		user, err := dbQueries.CreateUser(r.Context(), database.CreateUserParams{
			ID: pgtype.UUID{
				Bytes:  uuid.New(),
				Valid: true,
			},
			Username: req.Username,
		})
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	r.Get("/ws/{roomID}", func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "roomID")
		log.Printf("Client is connecting to room %s", roomID)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}
		
		hub.mu.Lock()
		hub.clients[conn] = true
		hub.mu.Unlock()

		go func() {
			defer func() {
				hub.mu.Lock()
				delete(hub.clients, conn)
				hub.mu.Unlock()
				conn.Close()
			}()
			
			for {
				// Read a message from the client
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("Client disconnected: %v", err)
					break // Exit the loop on error
				}

				// Place the message on the broadcast channel
				if messageType == websocket.TextMessage {
					log.Printf("Received message: %s", string(message))
					hub.broadcast <- message
				}
			}
		}()
	})

	// Server Start
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}

// run() is a goroutine that handles broadcasting messages to all clients.
func run() {
	for {
		// Wait for a message from the broadcast channel
		message := <-hub.broadcast
		
		// Send the message to every client in the map
		hub.mu.RLock()
		for client := range hub.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Write error to client: %v", err)
				client.Close()
				hub.mu.RUnlock() // unlock before deleting
				hub.mu.Lock()
				delete(hub.clients, client)
				hub.mu.Unlock()
				hub.mu.RLock()
			}
		}
		hub.mu.RUnlock()
	}
}
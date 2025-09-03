package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/handler"
	customMiddleware "github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"

	"github.com/mxhdiqaim/go-chat-app/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Go Chat Application API
// @version 1.0
// @description This is a chat application backend API.
// @host localhost:8080
// @BasePath /

func main() {
	// Load .env file. This should be the first thing in main.
    err := godotenv.Load()
    if err != nil {
        // This is not a fatal error, as in production we use real env vars.
        log.Println("Warning: .env file not found")
    }
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
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
	roomHandler := handler.NewRoomHandler(dbQueries, dbPool)
	userHandler := handler.NewUserHandler(dbQueries)

	hub := service.NewHub()
	go hub.Run()
	chatHandler := handler.NewChatHandler(hub, dbQueries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger Docs
	docs.SwaggerInfo.Title = "Go Chat Application API"
	docs.SwaggerInfo.Description = "This is a real-time chat application backend."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080" // Change this to your production URL when you deploy
	docs.SwaggerInfo.BasePath = "/"
    r.Get("/swagger/*", httpSwagger.Handler(
        httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // And here
    ))

	// Public Routes
	r.Post("/register", authHandler.RegisterUser)
	r.Post("/login", authHandler.LoginUser)

	// Protected Routes (with JWT middleware)
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)

		// User Endpoints
		r.Get("/users", userHandler.GetAllUsers)
		r.Get("/users/{id}", userHandler.GetUserByID)
		r.Get("/users/search", userHandler.SearchUsers)
		r.Put("/users/{id}", userHandler.UpdateUser)
		r.Delete("/users/{id}", userHandler.DeleteUser)

		// Room CRUD Endpoints
		r.Post("/rooms", roomHandler.CreateRoom)
		r.Get("/rooms", roomHandler.GetRooms)
		r.Get("/rooms/{id}", roomHandler.GetRoomByID)
		r.Put("/rooms/{id}", roomHandler.UpdateRoom)
		r.Delete("/rooms/{id}", roomHandler.DeleteRoom)
		r.Post("/rooms/{id}/join", roomHandler.JoinRoom)
		r.Delete("/rooms/{id}/leave", roomHandler.LeaveRoom)

		r.Get("/ws/{roomID}", chatHandler.ServeWs)
	})

	// Start Server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}
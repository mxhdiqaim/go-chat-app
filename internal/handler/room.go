package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
)

// RoomHandler handles requests related to chat rooms
type RoomHandler struct {
    db *database.Queries
    pool *pgxpool.Pool
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(db *database.Queries, pool *pgxpool.Pool) *RoomHandler {
    return &RoomHandler{db: db, pool: pool}
}

// CreateRoom handles creating a new room
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    var req struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusInternalServerError)
        return
    }

    // Start a transaction using the connection pool
    tx, err := h.pool.Begin(r.Context())
    if err != nil {
        http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback(r.Context()) // The defer will rollback if a commit doesn't happen

    // Create a new Queries object with the transaction
    qtx := database.New(tx)

    // 1. Create the room
    newRoom, err := qtx.CreateRoom(r.Context(), database.CreateRoomParams{
        Name:    req.Name,
        OwnerID: userUUID,
    })
    if err != nil {
        log.Println("Failed to create room:", err)
        http.Error(w, "Failed to create room", http.StatusInternalServerError)
        return
    }

    // 2. Add the owner as a member of the room
    err = qtx.AddRoomMember(r.Context(), database.AddRoomMemberParams{
        RoomID: newRoom.ID,
        UserID: userUUID,
    })
    if err != nil {
        log.Println("Failed to add room member:", err)
        http.Error(w, "Failed to add room member", http.StatusInternalServerError)
        return
    }

    // Commit the transaction if both operations succeeded
    if err := tx.Commit(r.Context()); err != nil {
        log.Println("Failed to commit transaction:", err)
        http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(newRoom)
}

// GetRooms gets all rooms
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
    rooms, err := h.db.GetRooms(r.Context())
    if err != nil {
        http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rooms)
}

// GetRoomByID gets a single room by its ID
func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
    roomIDParam := chi.URLParam(r, "id")
    roomID, err := uuid.Parse(roomIDParam)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    room, err := h.db.GetRoomByID(r.Context(), roomID)
    if err != nil {
        http.Error(w, "Room not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(room)
}

// UpdateRoom handles updating a room
func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
    roomIDParam := chi.URLParam(r, "id")
    roomID, err := uuid.Parse(roomIDParam)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    // Get user ID from the JWT token in the context
    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }
    
    // Check if the authenticated user is the room owner
    room, err := h.db.GetRoomByID(r.Context(), roomID)
    if err != nil {
        http.Error(w, "Room not found", http.StatusNotFound)
        return
    }

    if room.OwnerID.String() != userID {
        http.Error(w, "Forbidden: You are not the owner of this room", http.StatusForbidden)
        return
    }

    // Decode the request body
    var req struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    updatedRoom, err := h.db.UpdateRoom(r.Context(), database.UpdateRoomParams{
        ID:   roomID,
        Name: req.Name,
    })
    if err != nil {
        http.Error(w, "Failed to update room", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedRoom)
}

// DeleteRoom handles deleting a room
func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
    roomIDParam := chi.URLParam(r, "id")
    roomID, err := uuid.Parse(roomIDParam)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    // Get user ID from the JWT token in the context
    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }
    
    // Check if the authenticated user is the room owner
    room, err := h.db.GetRoomByID(r.Context(), roomID)
    if err != nil {
        http.Error(w, "Room not found", http.StatusNotFound)
        return
    }

    if room.OwnerID.String() != userID {
        http.Error(w, "Forbidden: You are not the owner of this room", http.StatusForbidden)
        return
    }

    if err := h.db.DeleteRoom(r.Context(), roomID); err != nil {
        http.Error(w, "Failed to delete room", http.StatusInternalServerError)
        return
    }

    // Return 204 No Content on successful deletion
    w.WriteHeader(http.StatusNoContent)
}

// JoinRoom handles a user joining a room.
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
    roomIDParam := chi.URLParam(r, "id")
    roomID, err := uuid.Parse(roomIDParam)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusInternalServerError)
        return
    }

    err = h.db.AddRoomMember(r.Context(), database.AddRoomMemberParams{
        RoomID: roomID,
        UserID: userUUID,
    })
    if err != nil {
        http.Error(w, "Failed to join room", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// LeaveRoom handles a user leaving a room.
func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
    roomIDParam := chi.URLParam(r, "id")
    roomID, err := uuid.Parse(roomIDParam)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated", http.StatusUnauthorized)
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusInternalServerError)
        return
    }

    err = h.db.RemoveRoomMember(r.Context(), database.RemoveRoomMemberParams{
        RoomID: roomID,
        UserID: userUUID,
    })
    if err != nil {
        http.Error(w, "Failed to leave room", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
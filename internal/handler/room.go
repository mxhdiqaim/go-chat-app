package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
)

// RoomHandler handles requests related to chat rooms
type RoomHandler struct {
    db *database.Queries
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(db *database.Queries) *RoomHandler {
    return &RoomHandler{db: db}
}

// CreateRoom handles creating a new room
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    // Get the user ID from the request context
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

    ownerID, err := uuid.Parse(userID)
    if err != nil {
        http.Error(w, "Invalid user ID in token", http.StatusInternalServerError)
        return
    }

    room, err := h.db.CreateRoom(r.Context(), database.CreateRoomParams{
        ID:      uuid.New(),
        Name:    req.Name,
        OwnerID: ownerID,
    })
    if err != nil {
        http.Error(w, "Failed to create room", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(room)
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
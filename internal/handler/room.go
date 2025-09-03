package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

// CreateRoomRequest defines the request body for creating a room.
type CreateRoomRequest struct {
    Name string `json:"name" example:"General"`
}

// RoomResponse defines the public shape of a room object.
type RoomResponse struct {
    ID        uuid.UUID `json:"id" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
    Name      string    `json:"name" example:"General"`
    OwnerID   uuid.UUID `json:"owner_id" example:"b1c2d3e4-f5g6-7890-1234-567890abcdef"`
    CreatedAt time.Time `json:"created_at" example:"2025-09-03T12:00:00Z"`
}

// CreateRoom godoc
// @Summary      Create a new room
// @Description  Creates a new chat room. The authenticated user becomes the owner.
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        room  body      CreateRoomRequest  true  "Room Name"
// @Success      201   {object}  RoomResponse
// @Failure      400   {string}  string "Invalid request body"
// @Failure      401   {string}  string "User not authenticated"
// @Failure      500   {string}  string "Failed to create room"
// @Security     ApiKeyAuth
// @Router       /rooms [post]
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    // Get the authenticated user's ID from the context.
    userIDString, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User ID not found in context", http.StatusUnauthorized)
        return
    }
    ownerID, err := uuid.Parse(userIDString)
    if err != nil {
        http.Error(w, "Invalid user ID format", http.StatusBadRequest)
        return
    }

    // Decode the room name from the request body.
    var req struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    if req.Name == "" {
        http.Error(w, "Room name is required", http.StatusBadRequest)
        return
    }

    // Call the database to create the room with a NEW UUID.
    params := database.CreateRoomParams{
        ID:      uuid.New(),
        Name:    req.Name,
        OwnerID: ownerID,
    }

    room, err := h.db.CreateRoom(r.Context(), params)
    if err != nil {
        log.Printf("Failed to create room: %v", err)
        http.Error(w, "Failed to create room", http.StatusInternalServerError)
        return
    }

    // Respond with the newly created room.
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(room)
}

// GetRooms godoc
// @Summary      Get all rooms
// @Description  Retrieves a list of all available chat rooms.
// @Tags         rooms
// @Produce      json
// @Success      200  {array}   RoomResponse
// @Failure      500  {string}  string "Failed to get rooms"
// @Router       /rooms [get]
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
    rooms, err := h.db.GetRooms(r.Context())
    if err != nil {
        http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
        return
    }

     // Convert database models to response DTOs
    var responses []RoomResponse
    for _, room := range rooms {
        responses = append(responses, RoomResponse{
            ID:        room.ID,
            Name:      room.Name,
            OwnerID:   room.OwnerID,
            CreatedAt: room.CreatedAt.Time,
        })
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(responses)
}

// GetRoomByID godoc
// @Summary      Get a single room by ID
// @Description  Retrieves details for a specific chat room.
// @Tags         rooms
// @Produce      json
// @Param        id  path      string  true  "Room ID"
// @Success      200 {object}  RoomResponse
// @Failure      400 {string}  string "Invalid room ID"
// @Failure      404 {string}  string "Room not found"
// @Router       /rooms/{id} [get]
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

    response := RoomResponse{
        ID:        room.ID,
        Name:      room.Name,
        OwnerID:   room.OwnerID,
        CreatedAt: room.CreatedAt.Time,
    }

    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// UpdateRoom godoc
// @Summary      Update a room
// @Description  Updates the name of a room. Only the room owner can perform this action.
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Room ID"
// @Param        room  body      CreateRoomRequest  true  "New Room Name"
// @Success      200   {object}  RoomResponse
// @Failure      400   {string}  string "Invalid room ID or request body"
// @Failure      401   {string}  string "User not authenticated"
// @Failure      403   {string}  string "Forbidden: You are not the owner"
// @Failure      404   {string}  string "Room not found"
// @Failure      500   {string}  string "Failed to update room"
// @Security     ApiKeyAuth
// @Router       /rooms/{id} [put]
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

    var req CreateRoomRequest
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

    response := RoomResponse{
        ID:        updatedRoom.ID,
        Name:      updatedRoom.Name,
        OwnerID:   updatedRoom.OwnerID,
        CreatedAt: updatedRoom.CreatedAt.Time,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// DeleteRoom godoc
// @Summary      Delete a room
// @Description  Deletes a room. Only the room owner can perform this action.
// @Tags         rooms
// @Param        id  path      string  true  "Room ID"
// @Success      204 {string}  string  "No Content"
// @Failure      400 {string}  string  "Invalid room ID"
// @Failure      401 {string}  string  "User not authenticated"
// @Failure      403 {string}  string  "Forbidden: You are not the owner"
// @Failure      404 {string}  string  "Room not found"
// @Failure      500 {string}  string  "Failed to delete room"
// @Security     ApiKeyAuth
// @Router       /rooms/{id} [delete]
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

// JoinRoom godoc
// @Summary      Join a room
// @Description  Adds the authenticated user to a room's member list.
// @Tags         rooms
// @Param        id  path      string  true  "Room ID to join"
// @Success      204 {string}  string  "No Content"
// @Failure      400 {string}  string  "Invalid room ID"
// @Failure      401 {string}  string  "User not authenticated"
// @Failure      500 {string}  string  "Failed to join room"
// @Security     ApiKeyAuth
// @Router       /rooms/{id}/join [post]
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

// LeaveRoom godoc
// @Summary      Leave a room
// @Description  Removes the authenticated user from a room's member list.
// @Tags         rooms
// @Param        id  path      string  true  "Room ID to leave"
// @Success      204 {string}  string  "No Content"
// @Failure      400 {string}  string  "Invalid room ID"
// @Failure      401 {string}  string  "User not authenticated"
// @Failure      500 {string}  string  "Failed to leave room"
// @Security     ApiKeyAuth
// @Router       /rooms/{id}/leave [post]
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
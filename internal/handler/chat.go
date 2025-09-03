package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mxhdiqaim/go-chat-app/internal/database"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

// ChatHandler handles the WebSocket endpoint.
type ChatHandler struct {
    hub *service.Hub
    db  *database.Queries
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(hub *service.Hub, db *database.Queries) *ChatHandler {
    return &ChatHandler{hub: hub, db: db}
}

// ServeWs godoc
// @Summary      Join and connect to a chat room
// @Description  Upgrades the HTTP connection to a WebSocket connection for a specific chat room. The user must be authenticated and a member of the room.
// @Tags         chat
// @Param        roomID  path      string  true  "Room ID to connect to"
// @Success      101     {string}  string  "Switching Protocols"
// @Failure      400     {string}  string  "Invalid room ID"
// @Failure      401     {string}  string  "User not authenticated"
// @Failure      403     {string}  string  "User is not a member of this room"
// @Failure      500     {string}  string  "Internal server error or failed to upgrade connection"
// @Security     ApiKeyAuth
// @Router       /ws/{roomID} [get]
func (h *ChatHandler) ServeWs(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated for WebSocket", http.StatusUnauthorized)
        return
    }

    roomID := chi.URLParam(r, "roomID")

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusInternalServerError)
        return
    }

    roomUUID, err := uuid.Parse(roomID)
    if err != nil {
        http.Error(w, "Invalid room ID", http.StatusBadRequest)
        return
    }

    isMember, err := h.db.IsRoomMember(r.Context(), database.IsRoomMemberParams{
        RoomID: roomUUID,
        UserID: userUUID,
    })
    if err != nil || !isMember {
        http.Error(w, "Forbidden: User is not a member of this room", http.StatusForbidden)
        return
    }

    conn, err := service.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    // Pass the roomID to the NewClient function
    client := service.NewClient(h.hub, conn, userID, roomID)
    client.Serve()
}
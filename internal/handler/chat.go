package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mxhdiqaim/go-chat-app/internal/middleware"
	"github.com/mxhdiqaim/go-chat-app/internal/service"
)

// ChatHandler handles the WebSocket endpoint.
type ChatHandler struct {
    hub *service.Hub
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(hub *service.Hub) *ChatHandler {
    return &ChatHandler{hub: hub}
}

// ServeWs handles websocket requests from the peer.
func (h *ChatHandler) ServeWs(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(middleware.ContextUserIDKey).(string)
    if !ok {
        http.Error(w, "User not authenticated for WebSocket", http.StatusUnauthorized)
        return
    }
	
    // We are no longer using roomID from URL param but from token
    _ = chi.URLParam(r, "roomID") // This is now ignored but we keep it for now

    conn, err := service.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }

    // Use the new constructor function to create the client
    client := service.NewClient(h.hub, conn, userID)

    // The Serve method is now on the client itself.
    client.Serve()
}
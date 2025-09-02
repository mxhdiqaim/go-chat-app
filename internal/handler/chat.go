package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	roomID := chi.URLParam(r, "roomID")
	log.Printf("New WebSocket connection for room: %s", roomID)

	// Use the exported Upgrader from the service package
	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

    client := service.NewClient(h.hub, conn)

	// client := &service.Client{hub: h.hub, conn: conn, send: make(chan []byte, 256)}
	// client.hub.register <- client

	// Start the read and write pumps
	client.Serve()
}
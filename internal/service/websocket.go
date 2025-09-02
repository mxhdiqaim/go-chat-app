package service

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
    // Registered clients for each room.
    clients map[string]map[string]*Client
    broadcast chan *Message
    register chan *Client
    unregister chan *Client
}

// Message represents a chat message.
type Message struct {
    SenderID    string `json:"sender_id"`
    RecipientID string `json:"recipient_id,omitempty"` // Omit if empty for broadcast messages
    RoomID      string `json:"room_id"`
    Content     string `json:"content"`
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
    hub *Hub
    conn *websocket.Conn
    send chan *Message
    userID string
    roomID string
}

// NewHub creates and returns a new Hub
func NewHub() *Hub {
    return &Hub{
        broadcast:  make(chan *Message),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        clients:    make(map[string]map[string]*Client),
    }
}


const (
    writeWait = 10 * time.Second
    pongWait = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
    maxMessageSize = 512
)

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            if _, ok := h.clients[client.roomID]; !ok {
                h.clients[client.roomID] = make(map[string]*Client)
            }
            h.clients[client.roomID][client.userID] = client
            log.Printf("Client %s registered to room %s", client.userID, client.roomID)

        case client := <-h.unregister:
            if _, ok := h.clients[client.roomID]; ok {
                if _, ok := h.clients[client.roomID][client.userID]; ok {
                    delete(h.clients[client.roomID], client.userID)
                    close(client.send)
                    log.Printf("Client %s unregistered from room %s", client.userID, client.roomID)
                }
            }
        case message := <-h.broadcast:
            if message.RecipientID != "" {
                if client, ok := h.clients[message.RoomID][message.RecipientID]; ok {
                    select {
                    case client.send <- message:
                    default:
                        close(client.send)
                        delete(h.clients[message.RoomID], client.userID)
                    }
                } else {
                    log.Printf("Recipient %s not found in room %s", message.RecipientID, message.RoomID)
                }
            } else {
                if clientsInRoom, ok := h.clients[message.RoomID]; ok {
                    for _, client := range clientsInRoom {
                        select {
                        case client.send <- message:
                        default:
                            close(client.send)
                            delete(h.clients[message.RoomID], client.userID)
                        }
                    }
                }
            }
        }
    }
}
// Upgrader exports the websocket upgrader for use in the handler package.
var Upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for development
    },
}

// NewClient creates a new client, registers it with the hub, and returns it.
func NewClient(hub *Hub, conn *websocket.Conn, userID, roomID string) *Client {
    client := &Client{
        hub:  hub,
        conn: conn,
        send: make(chan *Message, 256),
        userID: userID,
        roomID: roomID, // Initialize the new roomID field
    }
    client.hub.register <- client
    return client
}

// Serve handles the connection and starts the read and write pumps.
func (c *Client) Serve() {
    go c.readPump()
    go c.writePump()
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
    for {
        _, p, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }
        var message Message
        if err := json.Unmarshal(p, &message); err != nil {
            log.Printf("unmarshal error: %v", err)
            continue
        }
        message.SenderID = c.userID
        message.RoomID = c.roomID
        c.hub.broadcast <- &message
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            messageBytes, err := json.Marshal(message)
            if err != nil {
                log.Printf("json marshal error: %v", err)
                return
            }
            
            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(messageBytes)

            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                nextMessage := <-c.send
                nextMessageBytes, err := json.Marshal(nextMessage)
                if err != nil {
                    log.Printf("json marshal error: %v", err)
                    return
                }
                w.Write(nextMessageBytes)
            }
            if err := w.Close(); err != nil {
                return
            }
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
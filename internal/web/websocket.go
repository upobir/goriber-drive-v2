package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var ticker *time.Ticker

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	out  chan []byte
}

func (c *Client) close() {
	defer func() { recover() }()
	close(c.out)
	_ = c.conn.Close()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	ticker = time.NewTicker(45 * time.Second)
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
	}()

	for {
		select {
		case msg, ok := <-c.out:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		ev := NewWSEvent(EventWelcome, "Hello There")
		welcomeData, err := json.Marshal(ev)
		if err != nil {
			log.Println("marshal error:", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			out:  make(chan []byte, 64),
		}
		client.out <- welcomeData
		hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

type Hub struct {
	clients map[*Client]struct{}

	register   chan *Client
	unregister chan *Client

	broadcast chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for c := range h.clients {
				c.close()
			}
			return
		case c := <-h.register:
			h.clients[c] = struct{}{}
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.close()
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.out <- msg:
				default:
					delete(h.clients, c)
					c.close()
				}
			}
		}
	}
}

func (h *Hub) Broadcast(ev WSEvent) {
	data, _ := json.Marshal(ev)
	select {
	case h.broadcast <- data:
	default:
	}
}

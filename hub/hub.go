package hub

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ifonso/streaming-socket-wp/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Hub owns the set of clients and the last known state. Both are only
// touched by the Run goroutine; everything else talks to it through channels.
type Hub struct {
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan types.SpotifyPlayingState
	done       chan struct{}

	lastState types.SpotifyPlayingState
	hasState  bool
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan types.SpotifyPlayingState),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(h.done)
			for c := range h.clients {
				h.remove(c)
			}
			return
		case c := <-h.register:
			h.clients[c] = struct{}{}
			if h.hasState {
				c.send <- h.lastState
			}
		case c := <-h.unregister:
			h.remove(c)
		case state := <-h.broadcast:
			h.lastState = state
			h.hasState = true
			for c := range h.clients {
				select {
				case c.send <- state:
				default:
					// Client is not keeping up: drop it instead of blocking everyone.
					h.remove(c)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(state types.SpotifyPlayingState) {
	select {
	case h.broadcast <- state:
	case <-h.done:
	}
}

func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade: %v\n", err)
		return
	}

	c := newClient(h, conn)
	select {
	case h.register <- c:
	case <-h.done:
		conn.Close()
		return
	}

	go c.writePump()
	c.readPump()
}

func (h *Hub) unregisterClient(c *client) {
	select {
	case h.unregister <- c:
	case <-h.done:
	}
}

func (h *Hub) remove(c *client) {
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

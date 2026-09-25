package hub

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ifonso/streaming-socket-wp/types"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
	sendBufferSize = 8
)

// client is the only writer of its connection (writePump) and the only
// reader (readPump), as required by gorilla/websocket.
type client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan types.SpotifyPlayingState
}

func newClient(h *Hub, conn *websocket.Conn) *client {
	return &client{
		hub:  h,
		conn: conn,
		send: make(chan types.SpotifyPlayingState, sendBufferSize),
	}
}

func (c *client) writePump() {
	defer c.conn.Close()

	for state := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteJSON(state); err != nil {
			return
		}
	}

	// send was closed by the hub.
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *client) readPump() {
	defer func() {
		c.hub.unregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ReadMessage: %v\n", err)
			}
			return
		}
	}
}

package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	defaultPongWait = 60 * time.Second
	// Send pings at 90 % of pong-wait so the deadline is always refreshed in time.
	pingPeriod     = (defaultPongWait * 9) / 10
	maxMessageSize = 65536 // 64 KiB
)

// EventRouter is the interface the Client uses to dispatch incoming frames.
// It is defined as an interface here to break the import cycle between the
// websocket and handler packages.
type EventRouter interface {
	Route(ctx context.Context, c *Client, f *WSFrame) error
}

// Client wraps a single WebSocket connection and its send queue.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	userID       string
	chatIDs      []string
	router       EventRouter
	onDisconnect func()
	pongWait     time.Duration
}

// NewClient constructs a Client. The caller must call hub.Register(c) and
// start ReadPump / WritePump in separate goroutines.
func NewClient(hub *Hub, conn *websocket.Conn, userID string, chatIDs []string, router EventRouter) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		chatIDs:  chatIDs,
		router:   router,
		pongWait: defaultPongWait,
	}
}

// SetPongTimeout overrides the default pong-wait deadline (for tests).
func (c *Client) SetPongTimeout(d time.Duration) {
	c.pongWait = d
}

// Send queues a raw JSON payload for delivery. It is non-blocking; if the
// channel is full the message is silently dropped.
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
	}
}

// SendFrame marshals frame and queues it.
func (c *Client) SendFrame(frame *WSFrame) {
	data, err := json.Marshal(frame)
	if err != nil {
		log.Printf("websocket: failed to marshal frame: %v", err)
		return
	}
	c.Send(data)
}

// ReadPump pumps messages from the WebSocket connection to the event router.
// It runs until the connection is closed or an unrecoverable error occurs.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}()

	effective := c.pongWait
	pingInterval := (effective * 9) / 10

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(effective))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(effective))
		return nil
	})

	_ = pingInterval // used by WritePump

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return
		}

		var frame WSFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			// Disconnect on malformed JSON — protocol violation.
			return
		}

		if c.router != nil {
			if err := c.router.Route(context.Background(), c, &frame); err != nil {
				log.Printf("websocket route error (event=%s): %v", frame.Event, err)
			}
		}
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
// It also keeps the connection alive by sending periodic WebSocket pings.
func (c *Client) WritePump() {
	effective := c.pongWait
	interval := (effective * 9) / 10
	ticker := time.NewTicker(interval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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

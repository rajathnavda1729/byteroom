package websocket

import (
	"sync"
)

type broadcastMsg struct {
	chatID        string
	data          []byte
	exclude       *Client
	excludeUserID string // skip all connections for this user (e.g. HTTP sender)
}

// hubOp is a single serialized operation for the hub event loop. Using one
// channel guarantees FIFO ordering so a client is always registered in rooms
// before a concurrent HTTP broadcast is processed (Go's multi-channel select
// does not prioritize register over broadcast).
type hubOp interface {
	do(h *Hub)
}

type opRegister struct {
	c   *Client
	ack chan struct{} // if non-nil, closed after c is in rooms (WS handshake)
}

func (o opRegister) do(h *Hub) {
	h.mu.Lock()
	h.clients[o.c] = true
	for _, chatID := range o.c.chatIDs {
		if h.rooms[chatID] == nil {
			h.rooms[chatID] = make(map[*Client]bool)
		}
		h.rooms[chatID][o.c] = true
	}
	h.mu.Unlock()
	if o.ack != nil {
		close(o.ack)
	}
}

type opUnregister struct{ c *Client }

func (o opUnregister) do(h *Hub) {
	h.mu.Lock()
	if h.clients[o.c] {
		delete(h.clients, o.c)
		for _, chatID := range o.c.chatIDs {
			if room, ok := h.rooms[chatID]; ok {
				delete(room, o.c)
				if len(room) == 0 {
					delete(h.rooms, chatID)
				}
			}
		}
		close(o.c.send)
	}
	h.mu.Unlock()
}

type opBroadcast struct{ msg broadcastMsg }

func (o opBroadcast) do(h *Hub) {
	h.mu.RLock()
	room := h.rooms[o.msg.chatID]
	targets := make([]*Client, 0, len(room))
	for c := range room {
		if c == o.msg.exclude {
			continue
		}
		if o.msg.excludeUserID != "" && c.userID == o.msg.excludeUserID {
			continue
		}
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		select {
		case c.send <- o.msg.data:
		default:
			h.removeClient(c)
		}
	}
}

type opSubscribeUser struct{ userID, chatID string }

func (o opSubscribeUser) do(h *Hub) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if c.userID != o.userID {
			continue
		}
		if h.rooms[o.chatID] == nil {
			h.rooms[o.chatID] = make(map[*Client]bool)
		}
		h.rooms[o.chatID][c] = true
		for _, id := range c.chatIDs {
			if id == o.chatID {
				goto already
			}
		}
		c.chatIDs = append(c.chatIDs, o.chatID)
	already:
	}
}

type opBroadcastUser struct {
	userID string
	data   []byte
}

func (o opBroadcastUser) do(h *Hub) {
	h.mu.RLock()
	var targets []*Client
	for c := range h.clients {
		if c.userID == o.userID {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		select {
		case c.send <- o.data:
		default:
			h.removeClient(c)
		}
	}
}

// Hub manages all active WebSocket client connections and routes messages to
// the correct chat rooms. All state mutations are serialised through a single
// ops channel so ordering is deterministic.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	rooms   map[string]map[*Client]bool

	ops  chan hubOp
	done chan struct{}
	once sync.Once
}

// NewHub creates a Hub ready to be started with Run().
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		rooms:   make(map[string]map[*Client]bool),
		ops:     make(chan hubOp, 1024),
		done:    make(chan struct{}),
	}
}

// Run starts the hub event loop. Call it in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case op := <-h.ops:
			op.do(h)
		case <-h.done:
			return
		}
	}
}

// Register queues a client for registration with the hub.
func (h *Hub) Register(c *Client) {
	h.ops <- opRegister{c: c, ack: nil}
}

// RegisterWithAck registers the client and closes ack after they are in all
// initial chat rooms. The WS handler must wait on ack before relying on
// delivery for that connection.
func (h *Hub) RegisterWithAck(c *Client, ack chan struct{}) {
	h.ops <- opRegister{c: c, ack: ack}
}

// Unregister queues a client for removal from the hub.
func (h *Hub) Unregister(c *Client) {
	h.ops <- opUnregister{c}
}

// BroadcastToChat sends data to all members of chatID, optionally excluding
// one client (typically the sender).
func (h *Hub) BroadcastToChat(chatID string, data []byte, exclude *Client) {
	h.ops <- opBroadcast{broadcastMsg{chatID: chatID, data: data, exclude: exclude}}
}

// BroadcastToChatExceptUser sends data to all members of chatID except any
// connection belonging to excludeUserID (used when the sender used HTTP).
func (h *Hub) BroadcastToChatExceptUser(chatID string, data []byte, excludeUserID string) {
	h.ops <- opBroadcast{broadcastMsg{chatID: chatID, data: data, excludeUserID: excludeUserID}}
}

// HasClient reports whether the given client is currently registered.
func (h *Hub) HasClient(c *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[c]
}

// ActiveConnections returns the number of currently registered clients.
func (h *Hub) ActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SubscribeUserToRoom adds all currently-connected clients for userID to the
// given chat room so they receive future BroadcastToChat messages.
func (h *Hub) SubscribeUserToRoom(userID, chatID string) {
	h.ops <- opSubscribeUser{userID: userID, chatID: chatID}
}

// BroadcastToUser sends data to all connected clients belonging to userID.
func (h *Hub) BroadcastToUser(userID string, data []byte) {
	h.ops <- opBroadcastUser{userID: userID, data: data}
}

// Shutdown stops the hub's event loop.
func (h *Hub) Shutdown() {
	h.once.Do(func() { close(h.done) })
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	if h.clients[c] {
		delete(h.clients, c)
		for _, chatID := range c.chatIDs {
			if room, ok := h.rooms[chatID]; ok {
				delete(room, c)
				if len(room) == 0 {
					delete(h.rooms, chatID)
				}
			}
		}
		close(c.send)
	}
	h.mu.Unlock()
}

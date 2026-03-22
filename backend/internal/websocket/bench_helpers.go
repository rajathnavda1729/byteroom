package websocket

// NewBenchClient creates a minimal Client suitable for Hub benchmarks.
// The send channel is buffered; no real WebSocket connection is used.
func NewBenchClient(userID string, chatIDs []string) *Client {
	return &Client{
		userID:  userID,
		chatIDs: chatIDs,
		send:    make(chan []byte, 256),
	}
}

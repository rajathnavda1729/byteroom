package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_Run_RegistersClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	client := &Client{
		userID:  "user-123",
		chatIDs: []string{"chat-1"},
		send:    make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	assert.True(t, hub.HasClient(client))
}

func TestHub_Run_UnregistersClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	client := &Client{
		userID:  "user-123",
		chatIDs: []string{"chat-1"},
		send:    make(chan []byte, 256),
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)
	hub.Unregister(client)
	time.Sleep(10 * time.Millisecond)

	assert.False(t, hub.HasClient(client))
}

func TestHub_BroadcastToChat_DeliversToAllMembers(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	client1 := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	client2 := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	client3 := &Client{userID: "user-3", chatIDs: []string{"chat-2"}, send: make(chan []byte, 256)}

	hub.Register(client1)
	hub.Register(client2)
	hub.Register(client3)
	time.Sleep(10 * time.Millisecond)

	msg := []byte(`{"event":"message.new","data":{}}`)
	hub.BroadcastToChat("chat-1", msg, nil)

	select {
	case received := <-client1.send:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client1 did not receive message")
	}

	select {
	case received := <-client2.send:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("client2 did not receive message")
	}

	select {
	case <-client3.send:
		t.Fatal("client3 should not receive message")
	case <-time.After(50 * time.Millisecond):
		// Expected: client3 is not in chat-1
	}
}

func TestHub_BroadcastToChat_ExcludesSender(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}

	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	msg := []byte(`{"event":"message.new"}`)
	hub.BroadcastToChat("chat-1", msg, sender)

	select {
	case <-sender.send:
		t.Fatal("sender should not receive their own message")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	select {
	case <-receiver.send:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("receiver did not get message")
	}
}

func TestHub_ActiveConnections(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	assert.Equal(t, 0, hub.ActiveConnections())

	c1 := &Client{userID: "u1", chatIDs: []string{}, send: make(chan []byte, 1)}
	c2 := &Client{userID: "u2", chatIDs: []string{}, send: make(chan []byte, 1)}
	hub.Register(c1)
	hub.Register(c2)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 2, hub.ActiveConnections())
}

func TestHub_Shutdown_IsIdempotent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	hub.Shutdown()
	hub.Shutdown() // Must not panic
}

func TestHub_BroadcastToChat_NoRoom_DoesNothing(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Broadcast to a chatID with no registered clients — must not block/panic.
	hub.BroadcastToChat("nonexistent-chat", []byte(`{}`), nil)
	time.Sleep(20 * time.Millisecond) // Give the hub time to process
}

package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTypingHandler_Start_BroadcastsToChat(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := NewTypingHandler(hub)

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event: EventTypingStart,
		Data:  map[string]interface{}{"chat_id": "chat-1"},
	}

	handler.HandleStart(sender, frame)

	select {
	case msg := <-receiver.send:
		var typingFrame WSFrame
		json.Unmarshal(msg, &typingFrame)
		assert.Equal(t, EventUserTyping, typingFrame.Event)
		assert.Equal(t, true, typingFrame.Data["is_typing"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("receiver did not receive typing indicator")
	}

	// Sender should not receive their own typing indicator
	select {
	case <-sender.send:
		t.Fatal("sender should not receive own typing indicator")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestTypingHandler_Stop_BroadcastsStopToChat(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := NewTypingHandler(hub)

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event: EventTypingStop,
		Data:  map[string]interface{}{"chat_id": "chat-1"},
	}

	handler.HandleStop(sender, frame)

	select {
	case msg := <-receiver.send:
		var typingFrame WSFrame
		json.Unmarshal(msg, &typingFrame)
		assert.Equal(t, EventUserTyping, typingFrame.Event)
		assert.Equal(t, false, typingFrame.Data["is_typing"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("receiver did not receive typing stop")
	}
}

func TestTypingHandler_AutoTimeout_StopsAfter5Seconds(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := NewTypingHandler(hub)
	handler.SetTimeout(100 * time.Millisecond)

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event: EventTypingStart,
		Data:  map[string]interface{}{"chat_id": "chat-1"},
	}

	handler.HandleStart(sender, frame)
	<-receiver.send // Consume initial typing indicator

	// Wait for auto-timeout broadcast
	select {
	case msg := <-receiver.send:
		var typingFrame WSFrame
		json.Unmarshal(msg, &typingFrame)
		assert.Equal(t, EventUserTyping, typingFrame.Event)
		assert.Equal(t, false, typingFrame.Data["is_typing"])
	case <-time.After(500 * time.Millisecond):
		t.Fatal("typing did not auto-stop after timeout")
	}
}

func TestTypingHandler_Stop_CancelsAutoTimeout(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	handler := NewTypingHandler(hub)
	handler.SetTimeout(200 * time.Millisecond)

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	startFrame := &WSFrame{Event: EventTypingStart, Data: map[string]interface{}{"chat_id": "chat-1"}}
	stopFrame := &WSFrame{Event: EventTypingStop, Data: map[string]interface{}{"chat_id": "chat-1"}}

	handler.HandleStart(sender, startFrame)
	<-receiver.send // Consume start indicator

	// Explicit stop before timeout fires
	handler.HandleStop(sender, stopFrame)
	<-receiver.send // Consume stop indicator

	// No second stop should arrive from the auto-timeout
	select {
	case <-receiver.send:
		t.Fatal("auto-timeout fired after explicit stop")
	case <-time.After(400 * time.Millisecond):
		// Expected: timer was cancelled
	}
}

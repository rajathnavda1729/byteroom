package websocket_test

import (
	"fmt"
	"testing"
	"time"

	ws "github.com/byteroom/backend/internal/websocket"
)

func BenchmarkHub_BroadcastToChat(b *testing.B) {
	hub := ws.NewHub()
	go hub.Run()

	// Add 100 clients to chat-1
	for i := range 100 {
		client := ws.NewBenchClient(fmt.Sprintf("user-%d", i), []string{"chat-1"})
		hub.Register(client)
	}
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"event":"message.new","data":{"content":"test"}}`)

	b.ResetTimer()
	for range b.N {
		hub.BroadcastToChat("chat-1", msg, nil)
	}

	b.StopTimer()
	hub.Shutdown()
}

func BenchmarkHub_ConcurrentBroadcasts(b *testing.B) {
	hub := ws.NewHub()
	go hub.Run()

	// 10 chats × 50 clients
	for chat := range 10 {
		chatID := fmt.Sprintf("chat-%d", chat)
		for i := range 50 {
			client := ws.NewBenchClient(fmt.Sprintf("user-%d-%d", chat, i), []string{chatID})
			hub.Register(client)
		}
	}
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		msg := []byte(`{"event":"message.new","data":{}}`)
		i := 0
		for pb.Next() {
			chatID := fmt.Sprintf("chat-%d", i%10)
			hub.BroadcastToChat(chatID, msg, nil)
			i++
		}
	})

	b.StopTimer()
	hub.Shutdown()
}

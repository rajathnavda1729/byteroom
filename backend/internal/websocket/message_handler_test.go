package websocket

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageService satisfies the MessageService interface.
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) Send(ctx context.Context, msg *message.Message) (*message.Message, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*message.Message), args.Error(1)
}

func TestMessageHandler_HandleSend_ValidMessage_BroadcastsAndAcks(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	mockSvc := new(MockMessageService)
	mockSvc.On("Send", mock.Anything, mock.Anything).Return(&message.Message{
		ID:          "msg-abc",
		ChatID:      "chat-1",
		SenderID:    "user-1",
		ContentType: message.ContentTypeMarkdown,
		Content:     "Hello world",
		CreatedAt:   time.Now().UTC(),
	}, nil)

	handler := NewMessageEventHandler(hub, mockSvc)

	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	hub.Register(receiver)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event:     EventMessageSend,
		RequestID: "req-123",
		Data: map[string]interface{}{
			"message_id":   "msg-abc",
			"chat_id":      "chat-1",
			"content_type": "markdown",
			"content":      "Hello world",
		},
	}

	err := handler.HandleSend(context.Background(), sender, frame)
	assert.NoError(t, err)

	// ACK sent to sender
	select {
	case ack := <-sender.send:
		var ackFrame WSFrame
		json.Unmarshal(ack, &ackFrame)
		assert.Equal(t, EventMessageAck, ackFrame.Event)
		assert.Equal(t, "req-123", ackFrame.RequestID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sender did not receive ACK")
	}

	// Broadcast to receiver
	select {
	case msg := <-receiver.send:
		var msgFrame WSFrame
		json.Unmarshal(msg, &msgFrame)
		assert.Equal(t, EventMessageNew, msgFrame.Event)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("receiver did not receive message")
	}
}

func TestMessageHandler_HandleSend_NonMember_ReturnsError(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	mockSvc := new(MockMessageService)
	mockSvc.On("Send", mock.Anything, mock.Anything).Return(nil, message.ErrForbidden)

	handler := NewMessageEventHandler(hub, mockSvc)
	sender := &Client{userID: "user-1", chatIDs: []string{}, send: make(chan []byte, 256)}
	hub.Register(sender)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event: EventMessageSend,
		Data: map[string]interface{}{
			"chat_id": "chat-1",
			"content": "test",
		},
	}

	err := handler.HandleSend(context.Background(), sender, frame)
	assert.ErrorIs(t, err, message.ErrForbidden)

	// Error frame sent to sender
	select {
	case errMsg := <-sender.send:
		var errFrame WSFrame
		json.Unmarshal(errMsg, &errFrame)
		assert.Equal(t, EventMessageError, errFrame.Event)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sender did not receive error frame")
	}
}

func TestMessageHandler_HandleSend_DuplicateMessage_IdempotentAck(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	saved := &message.Message{
		ID:          "msg-duplicate",
		ChatID:      "chat-1",
		SenderID:    "user-1",
		ContentType: message.ContentTypeMarkdown,
		Content:     "test",
		CreatedAt:   time.Now().UTC(),
	}
	mockSvc := new(MockMessageService)
	mockSvc.On("Send", mock.Anything, mock.Anything).Return(saved, nil)

	handler := NewMessageEventHandler(hub, mockSvc)
	sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
	hub.Register(sender)
	time.Sleep(10 * time.Millisecond)

	frame := &WSFrame{
		Event: EventMessageSend,
		Data: map[string]interface{}{
			"message_id": "msg-duplicate",
			"chat_id":    "chat-1",
			"content":    "test",
		},
	}

	handler.HandleSend(context.Background(), sender, frame)
	// Drain ACK
	<-sender.send

	handler.HandleSend(context.Background(), sender, frame)
	// Both calls should succeed (idempotent — service handles dedup)
	mockSvc.AssertNumberOfCalls(t, "Send", 2)
}

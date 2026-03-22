package websocket

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter_Route_MessageSend_CallsHandler(t *testing.T) {
	called := false
	router := NewEventRouter()
	router.RegisterHandler(EventMessageSend, func(ctx context.Context, c *Client, f *WSFrame) error {
		called = true
		return nil
	})

	client := &Client{userID: "user-1", send: make(chan []byte, 256)}
	frame := &WSFrame{Event: EventMessageSend, Data: map[string]interface{}{}}

	err := router.Route(context.Background(), client, frame)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRouter_Route_UnknownEvent_ReturnsError(t *testing.T) {
	router := NewEventRouter()

	client := &Client{userID: "user-1", send: make(chan []byte, 256)}
	frame := &WSFrame{Event: "unknown.event", Data: map[string]interface{}{}}

	err := router.Route(context.Background(), client, frame)

	assert.ErrorIs(t, err, ErrUnknownEvent)
}

func TestRouter_Route_AllEvents_Dispatched(t *testing.T) {
	events := []string{EventMessageSend, EventTypingStart, EventTypingStop, EventPing}

	for _, event := range events {
		event := event
		t.Run(event, func(t *testing.T) {
			called := false
			router := NewEventRouter()
			router.RegisterHandler(event, func(ctx context.Context, c *Client, f *WSFrame) error {
				called = true
				return nil
			})

			client := &Client{userID: "user-1"}
			frame := &WSFrame{Event: event}

			err := router.Route(context.Background(), client, frame)

			// Route to an unknown event sends to c.send; skip that assertion for
			// events that ARE registered — they should succeed.
			assert.NoError(t, err)
			assert.True(t, called, "handler for %s not called", event)
		})
	}
}

func TestRouter_RegisterHandler_OverwritesExisting(t *testing.T) {
	router := NewEventRouter()

	calls := 0
	router.RegisterHandler(EventPing, func(ctx context.Context, c *Client, f *WSFrame) error {
		calls++
		return nil
	})
	router.RegisterHandler(EventPing, func(ctx context.Context, c *Client, f *WSFrame) error {
		calls += 10
		return nil
	})

	client := &Client{userID: "u1"}
	router.Route(context.Background(), client, &WSFrame{Event: EventPing})

	assert.Equal(t, 10, calls, "second registration should overwrite first")
}

package websocket

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// ErrUnknownEvent is returned by Router.Route when no handler is registered for
// the event contained in the frame.
var ErrUnknownEvent = errors.New("unknown websocket event")

// HandlerFunc is the function signature expected by the event router.
type HandlerFunc func(ctx context.Context, c *Client, f *WSFrame) error

// Router dispatches incoming WebSocket frames to the appropriate handler based
// on the event name. It implements the EventRouter interface used by Client.
type Router struct {
	handlers map[string]HandlerFunc
}

// NewEventRouter returns an empty Router.
func NewEventRouter() *Router {
	return &Router{handlers: make(map[string]HandlerFunc)}
}

// RegisterHandler associates a HandlerFunc with an event name.
func (r *Router) RegisterHandler(event string, h HandlerFunc) {
	r.handlers[event] = h
}

// Route looks up the handler for frame.Event and calls it.
// Returns ErrUnknownEvent when no handler is registered.
func (r *Router) Route(ctx context.Context, c *Client, f *WSFrame) error {
	h, ok := r.handlers[f.Event]
	if !ok {
		log.Printf("websocket: no handler for event %q", f.Event)
		c.SendFrame(&WSFrame{
			Event:     EventMessageError,
			RequestID: f.RequestID,
			Data: map[string]interface{}{
				"error": fmt.Sprintf("unknown event: %s", f.Event),
			},
		})
		return ErrUnknownEvent
	}
	return h(ctx, c, f)
}

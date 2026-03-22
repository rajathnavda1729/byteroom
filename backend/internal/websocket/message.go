package websocket

// WSFrame is the envelope for every WebSocket message exchanged with clients.
type WSFrame struct {
	Event     string                 `json:"event"`
	RequestID string                 `json:"request_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// Known event names.
const (
	EventMessageSend  = "message.send"
	EventMessageAck   = "message.ack"
	EventMessageNew   = "message.new"
	EventMessageError = "message.error"
	EventTypingStart  = "typing.start"
	EventTypingStop   = "typing.stop"
	EventUserTyping   = "user.typing"
	EventPing         = "ping"
	EventPong         = "pong"
)

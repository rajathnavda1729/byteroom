package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var testUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// pingRouter responds to "ping" events with "pong" so the heartbeat test can
// verify the application-level ping/pong round-trip.
type pingRouter struct{}

func (pr *pingRouter) Route(_ context.Context, c *Client, f *WSFrame) error {
	if f.Event == EventPing {
		c.SendFrame(&WSFrame{Event: EventPong, Data: map[string]interface{}{}})
	}
	return nil
}

func TestClient_ReadPump_ParsesValidMessage(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, "user-123", []string{"chat-1"}, nil)
		hub.Register(client)
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	msg := `{"event":"message.send","data":{"chat_id":"chat-1","content":"Hello"}}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
	assert.NoError(t, err)

	time.Sleep(20 * time.Millisecond)
}

func TestClient_WritePump_SendsQueuedMessages(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	done := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, "user-123", []string{"chat-1"}, nil)
		hub.Register(client)
		go client.WritePump()
		// Queue a message before ReadPump blocks
		client.Send([]byte(`{"event":"test"}`))
		client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	go func() {
		_, msg, _ := conn.ReadMessage()
		done <- msg
	}()

	select {
	case msg := <-done:
		assert.Contains(t, string(msg), "test")
	case <-time.After(time.Second):
		t.Fatal("did not receive queued message")
	}
}

func TestClient_ReadPump_DisconnectsOnInvalidJSON(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	disconnected := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, "user-123", []string{}, nil)
		client.onDisconnect = func() { close(disconnected) }
		hub.Register(client)
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))

	select {
	case <-disconnected:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("client should disconnect on invalid JSON")
	}
	conn.Close()
}

// TestClient_Heartbeat_ConnectionStaysAliveWithPongResponses verifies that when
// the browser client properly responds to WebSocket-level pings (gorilla
// handles this automatically), the server-side read deadline is continuously
// extended and the connection remains alive beyond a single timeout period.
//
// The connection alive-ness is verified by queuing a message through
// client.Send() after 2× the pong timeout and checking that the browser
// receives it — all writes go through WritePump to avoid concurrent write races.
func TestClient_Heartbeat_ConnectionStaysAliveWithPongResponses(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	var testClient *Client
	clientReady := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := NewClient(hub, conn, "user-123", []string{}, nil)
		// Short pong-wait so the test runs quickly.
		c.SetPongTimeout(150 * time.Millisecond)
		hub.Register(c)
		go c.WritePump()
		testClient = c
		close(clientReady)
		c.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Browser drains incoming messages and auto-responds to WebSocket pings
	received := make(chan []byte, 1)
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case received <- msg:
			default:
			}
		}
	}()

	// Wait until the test client is fully set up
	<-clientReady

	// After 2× the timeout, if the browser has been responding to pings the
	// connection is still open. Queue a probe message through WritePump.
	time.Sleep(350 * time.Millisecond)
	testClient.Send([]byte(`{"event":"alive"}`))

	select {
	case msg := <-received:
		assert.Contains(t, string(msg), "alive")
	case <-time.After(time.Second):
		t.Fatal("connection should remain alive while browser responds to pings")
	}
}

func TestClient_Heartbeat_DisconnectsOnPongTimeout(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	disconnected := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, "user-123", []string{}, nil)
		client.SetPongTimeout(100 * time.Millisecond)
		client.onDisconnect = func() { close(disconnected) }
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	// Ignore all pings so the server's read deadline expires
	conn.SetPingHandler(func(string) error { return nil })

	select {
	case <-disconnected:
		// Expected
	case <-time.After(time.Second):
		t.Fatal("client should have disconnected on pong timeout")
	}
	conn.Close()
}

func TestClient_Heartbeat_ClientPingGetsServerPong(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, "user-123", []string{}, &pingRouter{})
		hub.Register(client)
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"ping","data":{}}`))
	assert.NoError(t, err)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err)

	var frame WSFrame
	assert.NoError(t, json.Unmarshal(msg, &frame))
	assert.Equal(t, EventPong, frame.Event)
}

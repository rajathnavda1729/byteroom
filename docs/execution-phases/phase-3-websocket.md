# Phase 3: WebSocket & Real-time

## Objective

Implement the WebSocket infrastructure for real-time messaging: Hub architecture, client connection management, message routing, typing indicators, and presence system.

## Duration Estimate

5 development days

## Prerequisites

- Phase 2 completed
- All REST API endpoints functional
- Message service with idempotent persistence working
- JWT authentication validated

---

## Tasks

### Task 3.1: WebSocket Hub Core

**Description**: Implement the central Hub that manages all WebSocket connections and message routing.

**TDD Approach**:
```go
// internal/websocket/hub_test.go
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
```

**Subtasks**:
- [ ] Write Hub unit tests
- [ ] Implement `Hub` struct with channels for register/unregister/broadcast
- [ ] Implement `Run()` goroutine with select loop
- [ ] Implement client-to-chatRoom mapping
- [ ] Implement `BroadcastToChat()` with sender exclusion
- [ ] Add graceful shutdown with context cancellation
- [ ] Add metrics for active connections

**Exit Criteria**:
- [ ] All Hub tests pass
- [ ] Concurrent registration/unregistration is thread-safe
- [ ] Broadcasts reach all chat members
- [ ] Sender exclusion works correctly
- [ ] Coverage ≥ 90%

---

### Task 3.2: WebSocket Client Handler

**Description**: Implement the Client struct with read/write pumps for handling individual WebSocket connections.

**TDD Approach**:
```go
// internal/websocket/client_test.go
func TestClient_ReadPump_ParsesValidMessage(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    defer hub.Shutdown()
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        client := NewClient(hub, conn, "user-123", []string{"chat-1"})
        hub.Register(client)
        
        go client.WritePump()
        client.ReadPump() // Blocking
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    defer conn.Close()
    
    msg := `{"event":"message.send","data":{"chat_id":"chat-1","content":"Hello"}}`
    err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
    
    assert.NoError(t, err)
}

func TestClient_WritePump_SendsQueuedMessages(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    defer hub.Shutdown()
    
    var receivedMsg []byte
    done := make(chan bool)
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        client := NewClient(hub, conn, "user-123", []string{"chat-1"})
        hub.Register(client)
        
        go client.WritePump()
        
        // Queue a message
        client.Send([]byte(`{"event":"test"}`))
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    defer conn.Close()
    
    go func() {
        _, msg, _ := conn.ReadMessage()
        receivedMsg = msg
        done <- true
    }()
    
    select {
    case <-done:
        assert.Contains(t, string(receivedMsg), "test")
    case <-time.After(time.Second):
        t.Fatal("did not receive message")
    }
}

func TestClient_ReadPump_DisconnectsOnInvalidJSON(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    disconnected := make(chan bool)
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        client := NewClient(hub, conn, "user-123", []string{})
        client.onDisconnect = func() { disconnected <- true }
        hub.Register(client)
        go client.WritePump()
        client.ReadPump()
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    
    conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
    
    select {
    case <-disconnected:
        // Expected
    case <-time.After(time.Second):
        t.Fatal("client should disconnect on invalid JSON")
    }
}
```

**Subtasks**:
- [ ] Write Client tests
- [ ] Implement `Client` struct with send channel
- [ ] Implement `ReadPump()` with message parsing
- [ ] Implement `WritePump()` with ping/pong handling
- [ ] Handle connection close gracefully
- [ ] Implement message size limits (64KB)
- [ ] Add read/write deadlines

**Exit Criteria**:
- [ ] ReadPump correctly parses incoming messages
- [ ] WritePump delivers queued messages
- [ ] Ping/pong keeps connection alive
- [ ] Invalid messages handled gracefully
- [ ] Coverage ≥ 85%

---

### Task 3.3: WebSocket Upgrade Endpoint

**Description**: Implement the HTTP endpoint that upgrades connections to WebSocket with JWT authentication.

**TDD Approach**:
```go
// internal/api/handler/ws_handler_test.go
func TestWSHandler_Upgrade_ValidToken_Connects(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    jwtManager := NewJWTManager("secret", 24*time.Hour)
    chatRepo := new(MockChatRepository)
    chatRepo.On("GetUserChatIDs", mock.Anything, "user-123").Return([]string{"chat-1", "chat-2"}, nil)
    
    handler := NewWSHandler(hub, jwtManager, chatRepo)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    token, _ := jwtManager.Generate("user-123")
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token
    
    conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
    assert.NotNil(t, conn)
    conn.Close()
}

func TestWSHandler_Upgrade_InvalidToken_Returns401(t *testing.T) {
    hub := NewHub()
    jwtManager := NewJWTManager("secret", 24*time.Hour)
    handler := NewWSHandler(hub, jwtManager, nil)
    
    req := httptest.NewRequest("GET", "/ws?token=invalid", nil)
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWSHandler_Upgrade_MissingToken_Returns401(t *testing.T) {
    hub := NewHub()
    jwtManager := NewJWTManager("secret", 24*time.Hour)
    handler := NewWSHandler(hub, jwtManager, nil)
    
    req := httptest.NewRequest("GET", "/ws", nil)
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWSHandler_Upgrade_ExpiredToken_Returns401(t *testing.T) {
    hub := NewHub()
    jwtManager := NewJWTManager("secret", -1*time.Hour) // Expired
    handler := NewWSHandler(hub, jwtManager, nil)
    
    token, _ := jwtManager.Generate("user-123")
    req := httptest.NewRequest("GET", "/ws?token="+token, nil)
    rec := httptest.NewRecorder()
    
    handler.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

**Subtasks**:
- [ ] Write upgrade handler tests
- [ ] Implement `GET /ws` endpoint
- [ ] Extract and validate JWT from query param
- [ ] Fetch user's chat memberships
- [ ] Configure WebSocket upgrader (origin check, buffer sizes)
- [ ] Create Client and register with Hub
- [ ] Handle upgrade errors gracefully

**Exit Criteria**:
- [ ] Valid tokens establish WebSocket connection
- [ ] Invalid/expired tokens return 401
- [ ] Client registered with correct chat memberships
- [ ] Coverage ≥ 90%

---

### Task 3.4: Message Event Handler

**Description**: Implement the message.send event handler that processes incoming messages via WebSocket.

**TDD Approach**:
```go
// internal/websocket/handler_test.go
func TestMessageHandler_HandleSend_ValidMessage_BroadcastsAndAcks(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    mockMsgSvc := new(MockMessageService)
    mockMsgSvc.On("Send", mock.Anything, mock.Anything).Return(nil)
    
    handler := NewMessageHandler(hub, mockMsgSvc)
    
    sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    hub.Register(sender)
    hub.Register(receiver)
    time.Sleep(10 * time.Millisecond)
    
    frame := &WSFrame{
        Event:     "message.send",
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
    
    // Check ACK sent to sender
    select {
    case ack := <-sender.send:
        var ackFrame WSFrame
        json.Unmarshal(ack, &ackFrame)
        assert.Equal(t, "message.ack", ackFrame.Event)
        assert.Equal(t, "req-123", ackFrame.RequestID)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("sender did not receive ACK")
    }
    
    // Check broadcast to receiver
    select {
    case msg := <-receiver.send:
        var msgFrame WSFrame
        json.Unmarshal(msg, &msgFrame)
        assert.Equal(t, "message.new", msgFrame.Event)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("receiver did not receive message")
    }
}

func TestMessageHandler_HandleSend_NonMember_ReturnsError(t *testing.T) {
    hub := NewHub()
    mockMsgSvc := new(MockMessageService)
    mockMsgSvc.On("Send", mock.Anything, mock.Anything).Return(ErrForbidden)
    
    handler := NewMessageHandler(hub, mockMsgSvc)
    sender := &Client{userID: "user-1", chatIDs: []string{}, send: make(chan []byte, 256)}
    
    frame := &WSFrame{
        Event: "message.send",
        Data: map[string]interface{}{
            "chat_id": "chat-1",
            "content": "test",
        },
    }
    
    err := handler.HandleSend(context.Background(), sender, frame)
    
    assert.ErrorIs(t, err, ErrForbidden)
    
    // Check error sent to sender
    select {
    case errMsg := <-sender.send:
        var errFrame WSFrame
        json.Unmarshal(errMsg, &errFrame)
        assert.Equal(t, "message.error", errFrame.Event)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("sender did not receive error")
    }
}

func TestMessageHandler_HandleSend_DuplicateMessage_IdempotentAck(t *testing.T) {
    hub := NewHub()
    mockMsgSvc := new(MockMessageService)
    mockMsgSvc.On("Send", mock.Anything, mock.Anything).Return(nil) // Idempotent success
    
    handler := NewMessageHandler(hub, mockMsgSvc)
    sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    hub.Register(sender)
    
    frame := &WSFrame{
        Event: "message.send",
        Data: map[string]interface{}{
            "message_id": "msg-duplicate",
            "chat_id":    "chat-1",
            "content":    "test",
        },
    }
    
    // Send twice
    handler.HandleSend(context.Background(), sender, frame)
    handler.HandleSend(context.Background(), sender, frame)
    
    // Both should succeed (idempotent)
    mockMsgSvc.AssertNumberOfCalls(t, "Send", 2)
}
```

**Subtasks**:
- [ ] Write message handler tests
- [ ] Implement `HandleSend()` method
- [ ] Parse and validate message frame
- [ ] Call MessageService.Send()
- [ ] Send ACK to sender on success
- [ ] Send error frame on failure
- [ ] Broadcast message.new to chat members

**Exit Criteria**:
- [ ] Valid messages persisted and broadcast
- [ ] ACK sent with correct request_id
- [ ] Errors return message.error frame
- [ ] Idempotency handled correctly
- [ ] Coverage ≥ 85%

---

### Task 3.5: Typing Indicators

**Description**: Implement typing.start and typing.stop events with automatic timeout.

**TDD Approach**:
```go
// internal/websocket/typing_test.go
func TestTypingHandler_Start_BroadcastsToChat(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    handler := NewTypingHandler(hub)
    
    sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    hub.Register(sender)
    hub.Register(receiver)
    time.Sleep(10 * time.Millisecond)
    
    frame := &WSFrame{
        Event: "typing.start",
        Data:  map[string]interface{}{"chat_id": "chat-1"},
    }
    
    handler.HandleStart(sender, frame)
    
    select {
    case msg := <-receiver.send:
        var typingFrame WSFrame
        json.Unmarshal(msg, &typingFrame)
        assert.Equal(t, "user.typing", typingFrame.Event)
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
    
    handler := NewTypingHandler(hub)
    
    sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    hub.Register(sender)
    hub.Register(receiver)
    time.Sleep(10 * time.Millisecond)
    
    frame := &WSFrame{
        Event: "typing.stop",
        Data:  map[string]interface{}{"chat_id": "chat-1"},
    }
    
    handler.HandleStop(sender, frame)
    
    select {
    case msg := <-receiver.send:
        var typingFrame WSFrame
        json.Unmarshal(msg, &typingFrame)
        assert.Equal(t, "user.typing", typingFrame.Event)
        assert.Equal(t, false, typingFrame.Data["is_typing"])
    case <-time.After(100 * time.Millisecond):
        t.Fatal("receiver did not receive typing stop")
    }
}

func TestTypingHandler_AutoTimeout_StopsAfter5Seconds(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    handler := NewTypingHandler(hub)
    handler.SetTimeout(100 * time.Millisecond) // Short timeout for test
    
    sender := &Client{userID: "user-1", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    receiver := &Client{userID: "user-2", chatIDs: []string{"chat-1"}, send: make(chan []byte, 256)}
    hub.Register(sender)
    hub.Register(receiver)
    time.Sleep(10 * time.Millisecond)
    
    frame := &WSFrame{
        Event: "typing.start",
        Data:  map[string]interface{}{"chat_id": "chat-1"},
    }
    
    handler.HandleStart(sender, frame)
    <-receiver.send // Consume initial typing indicator
    
    // Wait for auto-timeout
    select {
    case msg := <-receiver.send:
        var typingFrame WSFrame
        json.Unmarshal(msg, &typingFrame)
        assert.Equal(t, false, typingFrame.Data["is_typing"])
    case <-time.After(200 * time.Millisecond):
        t.Fatal("typing did not auto-stop")
    }
}
```

**Subtasks**:
- [ ] Write typing handler tests
- [ ] Implement `HandleStart()` method
- [ ] Implement `HandleStop()` method
- [ ] Add typing state tracking per user per chat
- [ ] Implement auto-timeout (5 seconds)
- [ ] Cancel timeout on explicit stop or new message
- [ ] Broadcast user.typing event

**Exit Criteria**:
- [ ] Typing indicators broadcast to chat members
- [ ] Sender excluded from own typing events
- [ ] Auto-timeout after 5 seconds of inactivity
- [ ] Coverage ≥ 85%

---

### Task 3.6: Ping/Pong Heartbeat

**Description**: Implement heartbeat mechanism to detect stale connections and keep connections alive through proxies.

**TDD Approach**:
```go
// internal/websocket/heartbeat_test.go
func TestClient_Heartbeat_RespondsToServerPing(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    pongReceived := make(chan bool)
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        
        conn.SetPongHandler(func(string) error {
            pongReceived <- true
            return nil
        })
        
        client := NewClient(hub, conn, "user-123", []string{})
        go client.WritePump()
        go client.ReadPump()
        
        // Send ping
        conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    defer conn.Close()
    
    select {
    case <-pongReceived:
        // Expected
    case <-time.After(time.Second):
        t.Fatal("pong not received")
    }
}

func TestClient_Heartbeat_DisconnectsOnPongTimeout(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    disconnected := make(chan bool)
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        
        client := NewClient(hub, conn, "user-123", []string{})
        client.SetPongTimeout(100 * time.Millisecond)
        client.onDisconnect = func() { disconnected <- true }
        
        go client.WritePump()
        client.ReadPump()
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    
    // Don't respond to pings
    conn.SetPingHandler(func(string) error { return nil }) // Ignore
    
    select {
    case <-disconnected:
        // Expected - client should disconnect due to pong timeout
    case <-time.After(time.Second):
        t.Fatal("client should have disconnected")
    }
}

func TestClient_Heartbeat_ClientPingGetsServerPong(t *testing.T) {
    hub := NewHub()
    go hub.Run()
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        
        client := NewClient(hub, conn, "user-123", []string{})
        go client.WritePump()
        go client.ReadPump()
    }))
    defer server.Close()
    
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
    defer conn.Close()
    
    // Send ping event
    pingMsg := `{"event":"ping","data":{}}`
    conn.WriteMessage(websocket.TextMessage, []byte(pingMsg))
    
    _, msg, err := conn.ReadMessage()
    assert.NoError(t, err)
    
    var frame WSFrame
    json.Unmarshal(msg, &frame)
    assert.Equal(t, "pong", frame.Event)
}
```

**Subtasks**:
- [ ] Write heartbeat tests
- [ ] Configure ping interval (30 seconds)
- [ ] Configure pong timeout (60 seconds)
- [ ] Implement server-side ping in WritePump
- [ ] Handle pong responses with deadline extension
- [ ] Implement application-level ping/pong events
- [ ] Disconnect on pong timeout

**Exit Criteria**:
- [ ] Ping sent every 30 seconds
- [ ] Pong timeout disconnects stale clients
- [ ] Application ping/pong events work
- [ ] Coverage ≥ 90%

---

### Task 3.7: Event Router

**Description**: Implement the central event router that dispatches WebSocket events to appropriate handlers.

**TDD Approach**:
```go
// internal/websocket/router_test.go
func TestRouter_Route_MessageSend_CallsHandler(t *testing.T) {
    mockMsgHandler := new(MockMessageHandler)
    mockMsgHandler.On("HandleSend", mock.Anything, mock.Anything, mock.Anything).Return(nil)
    
    router := NewEventRouter()
    router.RegisterHandler("message.send", mockMsgHandler.HandleSend)
    
    client := &Client{userID: "user-1"}
    frame := &WSFrame{Event: "message.send", Data: map[string]interface{}{}}
    
    err := router.Route(context.Background(), client, frame)
    
    assert.NoError(t, err)
    mockMsgHandler.AssertCalled(t, "HandleSend", mock.Anything, client, frame)
}

func TestRouter_Route_UnknownEvent_ReturnsError(t *testing.T) {
    router := NewEventRouter()
    
    client := &Client{userID: "user-1", send: make(chan []byte, 256)}
    frame := &WSFrame{Event: "unknown.event", Data: map[string]interface{}{}}
    
    err := router.Route(context.Background(), client, frame)
    
    assert.ErrorIs(t, err, ErrUnknownEvent)
}

func TestRouter_Route_AllEvents_Dispatched(t *testing.T) {
    tests := []struct {
        event   string
        handler string
    }{
        {"message.send", "MessageHandler"},
        {"typing.start", "TypingHandler"},
        {"typing.stop", "TypingHandler"},
        {"ping", "PingHandler"},
    }
    
    for _, tt := range tests {
        t.Run(tt.event, func(t *testing.T) {
            called := false
            router := NewEventRouter()
            router.RegisterHandler(tt.event, func(ctx context.Context, c *Client, f *WSFrame) error {
                called = true
                return nil
            })
            
            client := &Client{userID: "user-1"}
            frame := &WSFrame{Event: tt.event}
            
            router.Route(context.Background(), client, frame)
            
            assert.True(t, called, "handler for %s not called", tt.event)
        })
    }
}
```

**Subtasks**:
- [ ] Write router tests
- [ ] Implement `EventRouter` struct
- [ ] Implement `RegisterHandler()` method
- [ ] Implement `Route()` method with event dispatch
- [ ] Add context propagation for request tracing
- [ ] Handle unknown events gracefully
- [ ] Add middleware support (logging, metrics)

**Exit Criteria**:
- [ ] All events routed to correct handlers
- [ ] Unknown events return error frame
- [ ] Context propagated through handlers
- [ ] Coverage ≥ 90%

---

### Task 3.8: Integration Tests

**Description**: Create end-to-end integration tests for the complete WebSocket flow.

**TDD Approach**:
```go
// internal/websocket/integration_test.go
// +build integration

func TestWebSocket_FullMessageFlow(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    hub := NewHub()
    go hub.Run()
    
    msgRepo := postgres.NewMessageRepository(db)
    chatRepo := postgres.NewChatRepository(db)
    userRepo := postgres.NewUserRepository(db)
    sanitizer := NewSanitizer()
    msgSvc := message.NewService(msgRepo, chatRepo, sanitizer, hub)
    jwtManager := NewJWTManager("test-secret", 24*time.Hour)
    
    // Create test data
    user1 := createTestUser(t, userRepo, "alice")
    user2 := createTestUser(t, userRepo, "bob")
    chat := createTestChat(t, chatRepo, user1.ID, user2.ID)
    
    handler := NewWSHandler(hub, jwtManager, chatRepo, msgSvc)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Connect user1
    token1, _ := jwtManager.Generate(user1.ID)
    ws1, _, _ := websocket.DefaultDialer.Dial(wsURL(server, token1), nil)
    defer ws1.Close()
    
    // Connect user2
    token2, _ := jwtManager.Generate(user2.ID)
    ws2, _, _ := websocket.DefaultDialer.Dial(wsURL(server, token2), nil)
    defer ws2.Close()
    
    // User1 sends message
    sendMsg := `{
        "event": "message.send",
        "request_id": "req-1",
        "data": {
            "message_id": "msg-test-1",
            "chat_id": "` + chat.ID + `",
            "content_type": "markdown",
            "content": "Hello from Alice!"
        }
    }`
    ws1.WriteMessage(websocket.TextMessage, []byte(sendMsg))
    
    // User1 receives ACK
    _, ackMsg, _ := ws1.ReadMessage()
    var ack WSFrame
    json.Unmarshal(ackMsg, &ack)
    assert.Equal(t, "message.ack", ack.Event)
    assert.Equal(t, "req-1", ack.RequestID)
    
    // User2 receives message
    _, newMsg, _ := ws2.ReadMessage()
    var received WSFrame
    json.Unmarshal(newMsg, &received)
    assert.Equal(t, "message.new", received.Event)
    assert.Equal(t, "Hello from Alice!", received.Data["content"])
    
    // Verify persisted in database
    msg, err := msgRepo.FindByID(context.Background(), "msg-test-1")
    assert.NoError(t, err)
    assert.Equal(t, "Hello from Alice!", msg.Content)
}

func TestWebSocket_TypingIndicators(t *testing.T) {
    // Setup similar to above...
    
    // User1 starts typing
    typingStart := `{
        "event": "typing.start",
        "data": {"chat_id": "` + chat.ID + `"}
    }`
    ws1.WriteMessage(websocket.TextMessage, []byte(typingStart))
    
    // User2 receives typing indicator
    _, typingMsg, _ := ws2.ReadMessage()
    var typing WSFrame
    json.Unmarshal(typingMsg, &typing)
    assert.Equal(t, "user.typing", typing.Event)
    assert.Equal(t, true, typing.Data["is_typing"])
    assert.Equal(t, user1.ID, typing.Data["user_id"])
}

func TestWebSocket_Reconnection_ReceivesMissedMessages(t *testing.T) {
    // Test that user receives messages from history on reconnect
    // (via REST API after reconnection)
}
```

**Subtasks**:
- [ ] Write full message flow integration test
- [ ] Write typing indicator integration test
- [ ] Write reconnection scenario test
- [ ] Write concurrent users test
- [ ] Test message idempotency under network conditions
- [ ] Test graceful disconnect and cleanup
- [ ] Add benchmarks for message throughput

**Exit Criteria**:
- [ ] All integration tests pass
- [ ] Message delivery ≤ 500ms latency
- [ ] No message loss under normal conditions
- [ ] 50 concurrent connections handled

---

## Phase 3 Exit Criteria Summary

### Automated Verification

```bash
# Run WebSocket tests
cd backend && go test ./internal/websocket/... -v -coverprofile=ws_coverage.out

# Run integration tests
cd backend && go test ./... -tags=integration -v

# Check coverage
go tool cover -func=ws_coverage.out | grep total
# Expected: >= 85%
```

### Manual Verification

| Check | Command | Expected Result |
|-------|---------|-----------------|
| WebSocket connects | `wscat -c "ws://localhost:8080/ws?token={token}"` | Connection established |
| Message delivery | Send message via WebSocket | ACK received, broadcast to members |
| Typing indicators | Send typing.start | Broadcast to chat members |
| Heartbeat | Wait 30+ seconds | Ping/pong keeps connection alive |

### Quality Gates

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] WebSocket coverage ≥ 85%
- [ ] Message latency < 500ms (p95)
- [ ] No race conditions (go test -race)
- [ ] Graceful shutdown works

### Deliverables

1. ✅ WebSocket Hub with broadcast capability
2. ✅ Client connection management
3. ✅ JWT-authenticated WebSocket upgrade
4. ✅ Message send/receive via WebSocket
5. ✅ Typing indicators
6. ✅ Heartbeat mechanism
7. ✅ Event router
8. ✅ Integration tests

---

## Next Phase

Upon completion, proceed to [Phase 4: Frontend](./phase-4-frontend.md)

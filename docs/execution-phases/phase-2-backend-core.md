# Phase 2: Backend Core Services

## Objective

Implement the core backend services: user management, authentication, chat management, and message handling (REST API only, WebSocket in Phase 3).

## Duration Estimate

7 development days

## Prerequisites

- Phase 1 completed
- Database running with migrations applied
- CI/CD pipeline functional

---

## Tasks

### Task 2.1: Configuration Management

**Description**: Implement configuration loading from environment variables.

**TDD Approach**:
```go
// internal/config/config_test.go
func TestConfig_Load_FromEnv(t *testing.T) {
    os.Setenv("PORT", "9090")
    os.Setenv("DB_HOST", "testhost")
    os.Setenv("JWT_SECRET", "testsecret")
    defer os.Clearenv()
    
    cfg, err := config.Load()
    
    assert.NoError(t, err)
    assert.Equal(t, 9090, cfg.Server.Port)
    assert.Equal(t, "testhost", cfg.Database.Host)
}

func TestConfig_Load_RequiredFieldMissing_ReturnsError(t *testing.T) {
    os.Clearenv()
    
    _, err := config.Load()
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "JWT_SECRET")
}
```

**Subtasks**:
- [ ] Write tests for configuration loading
- [ ] Implement `Config` struct with all fields
- [ ] Implement `Load()` function using `envconfig` or similar
- [ ] Add validation for required fields
- [ ] Create `.env.example` with all variables

**Exit Criteria**:
- [ ] All config tests pass
- [ ] Required fields validated
- [ ] Defaults work for optional fields
- [ ] Coverage ≥ 90% for config package

---

### Task 2.2: Database Connection Pool

**Description**: Implement PostgreSQL connection pool with health checks.

**TDD Approach**:
```go
// internal/infrastructure/postgres/db_test.go (integration)
// +build integration

func TestDB_Connect_ValidConfig_ReturnsPool(t *testing.T) {
    cfg := testConfig()
    
    pool, err := postgres.Connect(cfg)
    
    assert.NoError(t, err)
    assert.NotNil(t, pool)
    defer pool.Close()
}

func TestDB_Ping_HealthyConnection_ReturnsNil(t *testing.T) {
    pool := setupTestDB(t)
    defer pool.Close()
    
    err := pool.Ping(context.Background())
    
    assert.NoError(t, err)
}
```

**Subtasks**:
- [ ] Write connection tests (integration)
- [ ] Implement `Connect()` function
- [ ] Configure connection pool settings
- [ ] Implement graceful shutdown
- [ ] Add connection health check endpoint

**Exit Criteria**:
- [ ] Connection pool creates successfully
- [ ] Pool respects max connection limits
- [ ] Health check endpoint returns 200

---

### Task 2.3: User Domain - Entity & Repository

**Description**: Implement User entity and PostgreSQL repository.

**TDD Approach**:
```go
// internal/domain/user/repository_test.go
func TestUserRepository_Create_ValidUser_ReturnsID(t *testing.T) {
    repo := setupUserRepo(t)
    user := &User{
        Username:    "testuser",
        Email:       "test@example.com",
        PasswordHash: "hashed",
        DisplayName: "Test User",
    }
    
    id, err := repo.Create(context.Background(), user)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, id)
}

func TestUserRepository_Create_DuplicateEmail_ReturnsError(t *testing.T) {
    repo := setupUserRepo(t)
    user := &User{Email: "dup@example.com", Username: "user1", ...}
    repo.Create(context.Background(), user)
    
    user2 := &User{Email: "dup@example.com", Username: "user2", ...}
    _, err := repo.Create(context.Background(), user2)
    
    assert.ErrorIs(t, err, ErrDuplicate)
}

func TestUserRepository_FindByEmail_Exists_ReturnsUser(t *testing.T) {
    repo := setupUserRepo(t)
    created := createTestUser(t, repo)
    
    found, err := repo.FindByEmail(context.Background(), created.Email)
    
    assert.NoError(t, err)
    assert.Equal(t, created.ID, found.ID)
}

func TestUserRepository_FindByEmail_NotExists_ReturnsError(t *testing.T) {
    repo := setupUserRepo(t)
    
    _, err := repo.FindByEmail(context.Background(), "notfound@example.com")
    
    assert.ErrorIs(t, err, ErrNotFound)
}
```

**Subtasks**:
- [ ] Define `User` entity struct
- [ ] Define `UserRepository` interface
- [ ] Write repository tests (integration with testcontainers)
- [ ] Implement `PostgresUserRepository`
- [ ] Implement methods: `Create`, `FindByID`, `FindByEmail`, `FindByUsername`

**Exit Criteria**:
- [ ] All repository tests pass
- [ ] Duplicate detection works
- [ ] Proper error types returned
- [ ] Coverage ≥ 85%

---

### Task 2.4: User Service & Password Hashing

**Description**: Implement user service with secure password handling.

**TDD Approach**:
```go
// internal/domain/user/service_test.go
func TestUserService_Register_ValidInput_CreatesUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return("user-123", nil)
    svc := NewService(mockRepo)
    
    user, err := svc.Register(context.Background(), RegisterRequest{
        Username:    "newuser",
        Email:       "new@example.com",
        Password:    "SecurePass123!",
        DisplayName: "New User",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "user-123", user.ID)
    mockRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(u *User) bool {
        return u.PasswordHash != "SecurePass123!" // Password should be hashed
    }))
}

func TestUserService_Register_WeakPassword_ReturnsError(t *testing.T) {
    svc := NewService(nil) // Repo not needed, validation fails first
    
    _, err := svc.Register(context.Background(), RegisterRequest{
        Password: "weak",
    })
    
    assert.ErrorIs(t, err, ErrWeakPassword)
}

func TestUserService_Authenticate_ValidCredentials_ReturnsUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    hashedPw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(&User{
        ID:           "user-123",
        PasswordHash: string(hashedPw),
    }, nil)
    svc := NewService(mockRepo)
    
    user, err := svc.Authenticate(context.Background(), "test@example.com", "password123")
    
    assert.NoError(t, err)
    assert.Equal(t, "user-123", user.ID)
}
```

**Subtasks**:
- [ ] Write service tests with mocks
- [ ] Implement password validation rules
- [ ] Implement bcrypt hashing
- [ ] Implement `Register` method
- [ ] Implement `Authenticate` method
- [ ] Implement `GetByID` method

**Exit Criteria**:
- [ ] Passwords are bcrypt hashed (cost ≥ 10)
- [ ] Password validation enforces rules
- [ ] Authentication verifies hash correctly
- [ ] Coverage ≥ 90%

---

### Task 2.5: JWT Authentication

**Description**: Implement JWT token generation and validation.

**TDD Approach**:
```go
// internal/api/middleware/auth_test.go
func TestJWT_Generate_ValidUser_ReturnsToken(t *testing.T) {
    jwt := NewJWTManager("secret-key", 24*time.Hour)
    
    token, err := jwt.Generate("user-123")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestJWT_Validate_ValidToken_ReturnsUserID(t *testing.T) {
    jwt := NewJWTManager("secret-key", 24*time.Hour)
    token, _ := jwt.Generate("user-123")
    
    userID, err := jwt.Validate(token)
    
    assert.NoError(t, err)
    assert.Equal(t, "user-123", userID)
}

func TestJWT_Validate_ExpiredToken_ReturnsError(t *testing.T) {
    jwt := NewJWTManager("secret-key", -1*time.Hour) // Already expired
    token, _ := jwt.Generate("user-123")
    
    _, err := jwt.Validate(token)
    
    assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestAuthMiddleware_ValidToken_SetsUserContext(t *testing.T) {
    jwt := NewJWTManager("secret", 24*time.Hour)
    token, _ := jwt.Generate("user-123")
    middleware := AuthMiddleware(jwt)
    
    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rec := httptest.NewRecorder()
    
    handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value(UserIDKey).(string)
        assert.Equal(t, "user-123", userID)
    }))
    
    handler.ServeHTTP(rec, req)
}
```

**Subtasks**:
- [ ] Write JWT tests
- [ ] Implement `JWTManager` with generate/validate
- [ ] Implement refresh token logic
- [ ] Write auth middleware tests
- [ ] Implement `AuthMiddleware`
- [ ] Handle token expiration gracefully

**Exit Criteria**:
- [ ] Tokens contain user ID claim
- [ ] Expired tokens rejected
- [ ] Invalid tokens rejected
- [ ] Middleware sets user context
- [ ] Coverage ≥ 95%

---

### Task 2.6: Auth HTTP Handlers

**Description**: Implement `/api/auth/*` endpoints.

**TDD Approach**:
```go
// internal/api/handler/auth_handler_test.go
func TestAuthHandler_Register_ValidRequest_Returns201(t *testing.T) {
    mockSvc := new(MockUserService)
    mockSvc.On("Register", mock.Anything, mock.Anything).Return(&User{ID: "user-123"}, nil)
    handler := NewAuthHandler(mockSvc, jwtManager)
    
    body := `{"username":"test","email":"test@example.com","password":"Password123!","display_name":"Test"}`
    req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    
    handler.Register(rec, req)
    
    assert.Equal(t, http.StatusCreated, rec.Code)
    var resp RegisterResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    assert.NotEmpty(t, resp.Token)
}

func TestAuthHandler_Login_InvalidCredentials_Returns401(t *testing.T) {
    mockSvc := new(MockUserService)
    mockSvc.On("Authenticate", mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrInvalidCredentials)
    handler := NewAuthHandler(mockSvc, jwtManager)
    
    body := `{"email":"test@example.com","password":"wrong"}`
    req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
    rec := httptest.NewRecorder()
    
    handler.Login(rec, req)
    
    assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
```

**Subtasks**:
- [ ] Write handler tests
- [ ] Implement `POST /api/auth/register`
- [ ] Implement `POST /api/auth/login`
- [ ] Implement `POST /api/auth/refresh`
- [ ] Add request validation
- [ ] Add proper error responses

**Exit Criteria**:
- [ ] Registration creates user and returns token
- [ ] Login validates credentials and returns token
- [ ] Invalid requests return 400 with details
- [ ] Coverage ≥ 90%

---

### Task 2.7: Chat Domain - Entity & Repository

**Description**: Implement Chat and ChatMember entities with repository.

**TDD Approach**:
```go
// internal/domain/chat/repository_test.go
func TestChatRepository_Create_GroupChat_ReturnsID(t *testing.T) {
    repo := setupChatRepo(t)
    chat := &Chat{
        Name:      "Tech Discussion",
        Type:      ChatTypeGroup,
        CreatedBy: "user-123",
    }
    
    id, err := repo.Create(context.Background(), chat)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, id)
}

func TestChatRepository_AddMember_Success(t *testing.T) {
    repo := setupChatRepo(t)
    chatID := createTestChat(t, repo)
    
    err := repo.AddMember(context.Background(), chatID, "user-456", RoleMember)
    
    assert.NoError(t, err)
}

func TestChatRepository_IsMember_True(t *testing.T) {
    repo := setupChatRepo(t)
    chatID := createTestChat(t, repo)
    repo.AddMember(context.Background(), chatID, "user-456", RoleMember)
    
    isMember, err := repo.IsMember(context.Background(), chatID, "user-456")
    
    assert.NoError(t, err)
    assert.True(t, isMember)
}

func TestChatRepository_GetUserChats_ReturnsSortedByLastMessage(t *testing.T) {
    // Test that chats are ordered by most recent activity
}
```

**Subtasks**:
- [ ] Define `Chat`, `ChatMember` entities
- [ ] Define `ChatRepository` interface
- [ ] Write repository tests
- [ ] Implement `PostgresChatRepository`
- [ ] Methods: `Create`, `FindByID`, `AddMember`, `RemoveMember`, `IsMember`, `GetUserChats`

**Exit Criteria**:
- [ ] All repository tests pass
- [ ] Membership operations work correctly
- [ ] User's chats returned in correct order
- [ ] Coverage ≥ 85%

---

### Task 2.8: Chat Service & Handlers

**Description**: Implement chat service and REST endpoints.

**TDD Approach**:
```go
// internal/domain/chat/service_test.go
func TestChatService_CreateGroup_AddsCreatorAsAdmin(t *testing.T) {
    mockRepo := new(MockChatRepository)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return("chat-123", nil)
    mockRepo.On("AddMember", mock.Anything, "chat-123", "user-1", RoleAdmin).Return(nil)
    mockRepo.On("AddMember", mock.Anything, "chat-123", "user-2", RoleMember).Return(nil)
    svc := NewService(mockRepo)
    
    chat, err := svc.CreateGroup(context.Background(), "user-1", "My Group", []string{"user-2"})
    
    assert.NoError(t, err)
    assert.Equal(t, "chat-123", chat.ID)
}

func TestChatService_AddMember_NonAdmin_ReturnsError(t *testing.T) {
    mockRepo := new(MockChatRepository)
    mockRepo.On("GetMemberRole", mock.Anything, "chat-1", "user-1").Return(RoleMember, nil)
    svc := NewService(mockRepo)
    
    err := svc.AddMember(context.Background(), "chat-1", "user-1", "user-new")
    
    assert.ErrorIs(t, err, ErrForbidden)
}
```

**Subtasks**:
- [ ] Write service tests
- [ ] Implement `CreateDirectChat`, `CreateGroup`
- [ ] Implement `AddMember`, `RemoveMember`
- [ ] Write handler tests
- [ ] Implement `GET /api/chats`
- [ ] Implement `POST /api/chats`
- [ ] Implement `POST /api/chats/:id/members`
- [ ] Implement `DELETE /api/chats/:id/members/:userId`

**Exit Criteria**:
- [ ] Group creator becomes admin
- [ ] Only admins can add/remove members
- [ ] Direct chats have exactly 2 members
- [ ] Coverage ≥ 85%

---

### Task 2.9: Message Domain - Entity & Repository

**Description**: Implement Message entity with idempotent persistence.

**TDD Approach**:
```go
// internal/domain/message/repository_test.go
func TestMessageRepository_Save_NewMessage_Persists(t *testing.T) {
    repo := setupMessageRepo(t)
    msg := &Message{
        ID:          "msg-123",
        ChatID:      "chat-1",
        SenderID:    "user-1",
        ContentType: ContentTypeMarkdown,
        Content:     "Hello world",
    }
    
    err := repo.Save(context.Background(), msg)
    
    assert.NoError(t, err)
    
    // Verify persisted
    found, _ := repo.FindByID(context.Background(), "msg-123")
    assert.Equal(t, "Hello world", found.Content)
}

func TestMessageRepository_Save_DuplicateID_IsIdempotent(t *testing.T) {
    repo := setupMessageRepo(t)
    msg := &Message{ID: "msg-123", Content: "First"}
    repo.Save(context.Background(), msg)
    
    // Retry with same ID
    msg2 := &Message{ID: "msg-123", Content: "Second"}
    err := repo.Save(context.Background(), msg2)
    
    assert.NoError(t, err) // Should not error
    
    // Content should be original
    found, _ := repo.FindByID(context.Background(), "msg-123")
    assert.Equal(t, "First", found.Content)
}

func TestMessageRepository_FindByChatID_ReturnsPaginated(t *testing.T) {
    repo := setupMessageRepo(t)
    // Create 60 messages
    for i := 0; i < 60; i++ {
        repo.Save(context.Background(), &Message{ID: fmt.Sprintf("msg-%d", i), ChatID: "chat-1"})
    }
    
    msgs, err := repo.FindByChatID(context.Background(), "chat-1", 50, time.Now())
    
    assert.NoError(t, err)
    assert.Len(t, msgs, 50)
}
```

**Subtasks**:
- [ ] Define `Message` entity
- [ ] Define `MessageRepository` interface
- [ ] Write repository tests
- [ ] Implement `PostgresMessageRepository`
- [ ] Implement `Save` with ON CONFLICT DO NOTHING
- [ ] Implement `FindByID`, `FindByChatID` with cursor pagination

**Exit Criteria**:
- [ ] Duplicate saves are idempotent
- [ ] Pagination works correctly
- [ ] Messages ordered by timestamp DESC
- [ ] Coverage ≥ 90%

---

### Task 2.10: XSS Sanitizer

**Description**: Implement content sanitization using bluemonday.

**TDD Approach**:
```go
// internal/infrastructure/sanitizer/sanitizer_test.go
func TestSanitizer_Sanitize_RemovesScriptTags(t *testing.T) {
    s := NewSanitizer()
    
    input := `Hello <script>alert('xss')</script> world`
    result := s.Sanitize(input)
    
    assert.NotContains(t, result, "<script>")
    assert.Contains(t, result, "Hello")
    assert.Contains(t, result, "world")
}

func TestSanitizer_Sanitize_AllowsCodeBlocks(t *testing.T) {
    s := NewSanitizer()
    
    input := "```python\nprint('hello')\n```"
    result := s.Sanitize(input)
    
    assert.Contains(t, result, "print")
}

func TestSanitizer_Sanitize_RemovesOnClickHandlers(t *testing.T) {
    s := NewSanitizer()
    
    input := `<a href="#" onclick="evil()">Click</a>`
    result := s.Sanitize(input)
    
    assert.NotContains(t, result, "onclick")
}

func TestSanitizer_Sanitize_AllowsMarkdownFormatting(t *testing.T) {
    s := NewSanitizer()
    
    input := "**bold** and *italic* and `code`"
    result := s.Sanitize(input)
    
    assert.Equal(t, input, result) // Should pass through
}
```

**Subtasks**:
- [ ] Write sanitizer tests for attack vectors
- [ ] Implement `Sanitizer` interface
- [ ] Configure bluemonday policy
- [ ] Allow safe elements (code, pre, etc.)
- [ ] Block dangerous elements/attributes

**Exit Criteria**:
- [ ] Script tags removed
- [ ] Event handlers removed
- [ ] Safe markdown preserved
- [ ] Code blocks preserved
- [ ] Coverage ≥ 95%

---

### Task 2.11: Message Service & Handlers

**Description**: Implement message service and history endpoint.

**TDD Approach**:
```go
// internal/domain/message/service_test.go
func TestMessageService_Send_ValidMessage_Persists(t *testing.T) {
    mockRepo := new(MockMessageRepository)
    mockRepo.On("Exists", mock.Anything, "msg-1").Return(false, nil)
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
    mockChatRepo := new(MockChatRepository)
    mockChatRepo.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)
    mockSanitizer := new(MockSanitizer)
    mockSanitizer.On("Sanitize", mock.Anything).Return("cleaned content")
    
    svc := NewService(mockRepo, mockChatRepo, mockSanitizer, nil)
    
    err := svc.Send(context.Background(), &Message{
        ID:       "msg-1",
        ChatID:   "chat-1",
        SenderID: "user-1",
        Content:  "<script>bad</script>Hello",
    })
    
    assert.NoError(t, err)
    mockSanitizer.AssertCalled(t, "Sanitize", "<script>bad</script>Hello")
}

func TestMessageService_Send_NonMember_ReturnsError(t *testing.T) {
    mockChatRepo := new(MockChatRepository)
    mockChatRepo.On("IsMember", mock.Anything, "chat-1", "user-1").Return(false, nil)
    
    svc := NewService(nil, mockChatRepo, nil, nil)
    
    err := svc.Send(context.Background(), &Message{ChatID: "chat-1", SenderID: "user-1"})
    
    assert.ErrorIs(t, err, ErrForbidden)
}
```

**Subtasks**:
- [ ] Write service tests
- [ ] Implement `Send` with membership check and sanitization
- [ ] Implement `GetHistory` with pagination
- [ ] Write handler tests
- [ ] Implement `GET /api/chats/:id/messages`

**Exit Criteria**:
- [ ] Messages sanitized before save
- [ ] Non-members cannot send/view messages
- [ ] History paginated correctly
- [ ] Coverage ≥ 85%

---

### Task 2.12: S3 Pre-signed URL Handler

**Description**: Implement media upload flow with S3 pre-signed URLs.

**TDD Approach**:
```go
// internal/api/handler/upload_handler_test.go
func TestUploadHandler_RequestURL_ValidMimeType_Returns200(t *testing.T) {
    mockS3 := new(MockS3Client)
    mockS3.On("GeneratePresignedURL", mock.Anything, mock.Anything).Return("https://s3.../upload", nil)
    handler := NewUploadHandler(mockS3)
    
    body := `{"filename":"diagram.png","mime_type":"image/png","size_bytes":1024}`
    req := httptest.NewRequest("POST", "/api/upload/request", strings.NewReader(body))
    rec := httptest.NewRecorder()
    
    handler.RequestUploadURL(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
    var resp UploadURLResponse
    json.Unmarshal(rec.Body.Bytes(), &resp)
    assert.NotEmpty(t, resp.UploadURL)
    assert.NotEmpty(t, resp.FileKey)
}

func TestUploadHandler_RequestURL_InvalidMimeType_Returns400(t *testing.T) {
    handler := NewUploadHandler(nil)
    
    body := `{"filename":"virus.exe","mime_type":"application/x-executable","size_bytes":1024}`
    req := httptest.NewRequest("POST", "/api/upload/request", strings.NewReader(body))
    rec := httptest.NewRecorder()
    
    handler.RequestUploadURL(rec, req)
    
    assert.Equal(t, http.StatusBadRequest, rec.Code)
}
```

**Subtasks**:
- [ ] Write handler tests
- [ ] Implement S3 client wrapper
- [ ] Implement `POST /api/upload/request`
- [ ] Validate mime types (images only)
- [ ] Validate file size (max 10MB)
- [ ] Generate unique file keys

**Exit Criteria**:
- [ ] Pre-signed URLs generated correctly
- [ ] Invalid mime types rejected
- [ ] Oversized files rejected
- [ ] Coverage ≥ 85%

---

## Phase 2 Exit Criteria Summary

### Automated Verification

```bash
# Run all backend tests
cd backend && go test ./... -coverprofile=coverage.out

# Check coverage
go tool cover -func=coverage.out | grep total
# Expected: total: (statements) >= 80.0%
```

### API Endpoint Verification

| Endpoint | Method | Test Command |
|----------|--------|--------------|
| /api/auth/register | POST | `curl -X POST localhost:8080/api/auth/register -d '{"username":"test",...}'` |
| /api/auth/login | POST | `curl -X POST localhost:8080/api/auth/login -d '{"email":"..."}'` |
| /api/chats | GET | `curl -H "Authorization: Bearer {token}" localhost:8080/api/chats` |
| /api/chats | POST | `curl -X POST -H "Authorization: ..." localhost:8080/api/chats -d '{...}'` |
| /api/chats/:id/messages | GET | `curl -H "Authorization: ..." localhost:8080/api/chats/123/messages` |
| /api/upload/request | POST | `curl -X POST -H "Authorization: ..." localhost:8080/api/upload/request` |

### Quality Gates

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Code coverage ≥ 80%
- [ ] No critical security issues
- [ ] All endpoints return correct status codes
- [ ] Error responses follow API spec

### Deliverables

1. ✅ Configuration management
2. ✅ Database connection pool
3. ✅ User domain (entity, repository, service)
4. ✅ JWT authentication
5. ✅ Auth endpoints
6. ✅ Chat domain (entity, repository, service)
7. ✅ Chat endpoints
8. ✅ Message domain with idempotency
9. ✅ XSS sanitization
10. ✅ Message history endpoint
11. ✅ S3 upload flow

---

## Next Phase

Upon completion, proceed to [Phase 3: WebSocket & Real-time](./phase-3-websocket.md)

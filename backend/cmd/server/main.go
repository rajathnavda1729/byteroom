package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/byteroom/backend/internal/api"
	"github.com/byteroom/backend/internal/api/handler"
	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/config"
	"github.com/byteroom/backend/internal/domain/chat"
	"github.com/byteroom/backend/internal/domain/message"
	"github.com/byteroom/backend/internal/domain/user"
	"github.com/byteroom/backend/internal/infrastructure/postgres"
	"github.com/byteroom/backend/internal/infrastructure/sanitizer"
	wsinternal "github.com/byteroom/backend/internal/websocket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// Database
	db, err := postgres.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	msgRepo := postgres.NewMessageRepository(db)

	// Services
	userSvc := user.NewService(userRepo)
	chatSvc := chat.NewService(chatRepo)
	msgSvc := message.NewService(msgRepo, chatRepo, sanitizerAdapter{})

	// JWT
	jwtMgr := middleware.NewJWTManager(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpiryHours)*time.Hour)

	// WebSocket Hub
	hub := wsinternal.NewHub()
	go hub.Run()

	// WebSocket event router
	wsRouter := wsinternal.NewEventRouter()
	msgEventHandler := wsinternal.NewMessageEventHandler(hub, msgSvc)
	typingHandler := wsinternal.NewTypingHandler(hub)

	wsRouter.RegisterHandler(wsinternal.EventMessageSend, msgEventHandler.HandleSend)
	wsRouter.RegisterHandler(wsinternal.EventTypingStart, func(ctx context.Context, c *wsinternal.Client, f *wsinternal.WSFrame) error {
		typingHandler.HandleStart(c, f)
		return nil
	})
	wsRouter.RegisterHandler(wsinternal.EventTypingStop, func(ctx context.Context, c *wsinternal.Client, f *wsinternal.WSFrame) error {
		typingHandler.HandleStop(c, f)
		return nil
	})
	wsRouter.RegisterHandler(wsinternal.EventPing, func(_ context.Context, c *wsinternal.Client, f *wsinternal.WSFrame) error {
		c.SendFrame(&wsinternal.WSFrame{Event: wsinternal.EventPong, Data: map[string]interface{}{}})
		return nil
	})

	// HTTP Handlers
	authHandler := handler.NewAuthHandler(userSvc, jwtMgr)
	chatHandler := handler.NewChatHandler(chatSvc, chatRepo).WithNotifier(hub)
	msgHandler := handler.NewMessageHandler(msgSvc).WithHub(hub)
	uploadHandler := handler.NewUploadHandler(nil) // S3 client injected when configured
	healthHandler := handler.NewHealthHandler().WithDB(db).WithHub(hub)
	wsHandler := handler.NewWSHandler(hub, jwtMgr, chatRepo, wsRouter)

	// Router
	origins := []string{
		"http://localhost:5173",
		fmt.Sprintf("http://localhost:%s", cfg.Server.Port),
	}
	router := api.NewRouter(api.RouterConfig{
		Auth:    authHandler,
		Chats:   chatHandler,
		Msgs:    msgHandler,
		Uploads: uploadHandler,
		Health:  healthHandler,
		WS:      wsHandler,
		JWT:     jwtMgr,
		Origins: origins,
	})

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("ByteRoom server starting on %s (env=%s)", addr, cfg.Server.Env)

	srv := &http.Server{Addr: addr, Handler: router.Build()}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server…")
		hub.Shutdown()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
		os.Exit(1)
	}
	log.Println("Server stopped")
}

// sanitizerAdapter adapts the infrastructure sanitizer to the message.Sanitizer interface.
type sanitizerAdapter struct{}

func (s sanitizerAdapter) Sanitize(content string) string {
	return sanitizer.SanitizeMarkdown(content)
}

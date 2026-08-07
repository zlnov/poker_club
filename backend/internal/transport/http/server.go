package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/config"
	"poker-club/backend/internal/service"
)

// Server wraps the HTTP server and Gin engine.
type Server struct {
	cfg     *config.Config
	router  *gin.Engine
	httpSrv *http.Server
}

// NewServer creates a new HTTP server with routes and middleware configured.
func NewServer(cfg *config.Config, svc *service.Service) *Server {
	// Disable debug mode in production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware chain: Recovery -> Logger
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Healthcheck endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := svc.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Webhook endpoint placeholder — registered by transport layer
	router.POST("/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return &Server{
		cfg:    cfg,
		router: router,
	}
}

// RegisterWebhookHandler registers a custom handler for the /webhook path.
func (s *Server) RegisterWebhookHandler(handler func(c *gin.Context)) {
	s.router.POST("/webhook", handler)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:    s.cfg.WebhookAddr(),
		Handler: s.router,
	}

	return s.httpSrv.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.httpSrv.Shutdown(shutdownCtx)
}

// Router returns the Gin engine for adding routes.
func (s *Server) Router() *gin.Engine {
	return s.router
}

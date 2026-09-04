package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/auth"
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
func NewServer(cfg *config.Config, svc *service.Service, jwt *auth.JWTManager, authUC *auth.AuthUseCase) *Server {
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

	// Authentication routes (public)
	authHandler := NewAuthHandler(authUC)
	authRoutes := router.Group("/api/v1/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/telegram", authHandler.TelegramAuth)
		authRoutes.POST("/refresh", authHandler.Refresh)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	// Protected routes (require JWT authentication)
	protected := router.Group("/api/v1")
	protected.Use(AuthMiddleware(jwt))
	{
		protected.GET("/me", authHandler.Me)

		// Club management
		clubHandler := NewClubHandler(svc)
		protected.GET("/clubs", clubHandler.ListClubs)
		protected.POST("/clubs", clubHandler.CreateClub)
		protected.GET("/clubs/:clubId", clubHandler.GetClub)
		protected.PATCH("/clubs/:clubId", clubHandler.UpdateClub)
		protected.DELETE("/clubs/:clubId", clubHandler.CloseClub)

		// Club membership
		memberHandler := NewClubMemberHandler(svc)
		protected.GET("/clubs/:clubId/members", memberHandler.ListMembers)
		protected.POST("/clubs/:clubId/members", memberHandler.InviteMember)
		protected.PATCH("/clubs/:clubId/members/:playerId", memberHandler.UpdateMember)
		protected.DELETE("/clubs/:clubId/members/:playerId", memberHandler.RemoveMember)

		// Membership requests
		membershipHandler := NewMembershipRequestHandler(svc)
		protected.GET("/clubs/:clubId/member-requests", membershipHandler.ListMemberRequests)
		protected.POST("/clubs/:clubId/members/:playerId/approve", membershipHandler.ApproveMember)
		protected.POST("/clubs/:clubId/members/:playerId/reject", membershipHandler.RejectMember)

		// Club invitations
		invitationHandler := NewClubInvitationHandler(svc)
		protected.GET("/clubs/:clubId/invites", invitationHandler.ListInvites)
		protected.POST("/clubs/:clubId/invites", invitationHandler.CreateInvite)

		// Statistics
		statsHandler := NewStatisticsHandler(svc)
		protected.GET("/clubs/:clubId/statistics", statsHandler.GetClubStatistics)
		protected.GET("/players/:playerId/statistics", statsHandler.GetPlayerStatistics)

		// Games
		gameHandler := NewGameHandler(svc)
		protected.GET("/clubs/:clubId/games", gameHandler.ListGames)
		protected.POST("/clubs/:clubId/games", gameHandler.CreateGame)
		protected.GET("/games/:gameId", gameHandler.GetGame)
		protected.PATCH("/games/:gameId", gameHandler.UpdateGame)
		protected.POST("/games/:gameId/start", gameHandler.StartGame)
		protected.POST("/games/:gameId/finish", gameHandler.FinishGame)
		protected.POST("/games/:gameId/cancel", gameHandler.CancelGame)
		protected.PATCH("/games/:gameId/banker", gameHandler.UpdateBanker)
		protected.GET("/games/:gameId/participants", gameHandler.ListGameParticipants)
		protected.POST("/games/:gameId/participants", gameHandler.AddGameParticipant)
		protected.DELETE("/games/:gameId/participants/:playerId", gameHandler.RemoveGameParticipant)
		protected.POST("/games/:gameId/participants/:playerId/buy-in", gameHandler.RegisterBuyIn)
		protected.POST("/games/:gameId/participants/:playerId/rebuy", gameHandler.RegisterRebuy)
		protected.POST("/games/:gameId/participants/:playerId/chips", gameHandler.SetChipsEnd)
		protected.GET("/games/:gameId/results", gameHandler.GetGameResults)
		protected.POST("/games/:gameId/results/correct", gameHandler.CorrectGameResults)
		protected.GET("/games/:gameId/monitor", gameHandler.GetGameMonitor)
		protected.GET("/games/:gameId/events", gameHandler.GetGameEvents)
	}

	// Webhook endpoint — registered by transport layer via RegisterWebhookHandler

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

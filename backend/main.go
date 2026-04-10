package main

import (
	"log"

	"poker-club-backend/application/usecases"
	"poker-club-backend/domain"
	"poker-club-backend/infrastructure/persistence"
	"poker-club-backend/infrastructure/services"
	"poker-club-backend/presentation/handlers"
	"poker-club-backend/presentation/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := persistence.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to load env: %v", err)
	}

	// Connect to database
	db, err := persistence.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate tables (for development)
	if err := db.AutoMigrate(
		&persistence.Club{},
		&persistence.Player{},
		&persistence.ClubMember{},
		&persistence.Game{},
		&persistence.GameParticipant{},
		&persistence.Event{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	clubRepo := persistence.NewClubRepository(db)
	playerRepo := persistence.NewPlayerRepository(db)
	memberRepo := persistence.NewClubMemberRepository(db)
	gameRepo := persistence.NewGameRepository(db)
	participantRepo := persistence.NewGameParticipantRepository(db)
	eventRepo := persistence.NewEventRepository(db)

	// Initialize domain services
	clubService := domain.NewClubService(clubRepo, memberRepo, playerRepo)
	gameService := domain.NewGameService(gameRepo, participantRepo, eventRepo, memberRepo, playerRepo, clubRepo)

	// Initialize use cases
	clubUseCase := usecases.NewClubUseCase(clubService)
	gameUseCase := usecases.NewGameUseCase(gameService)

	// Initialize JWT service
	jwtService := services.NewJWTService(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
	)

	// Initialize auth use case
	authUseCase := usecases.NewAuthUseCase(playerRepo, jwtService)

	// Initialize handlers
	clubHandler := handlers.NewClubHandler(clubUseCase)
	gameHandler := handlers.NewGameHandler(gameUseCase)
	authHandler := handlers.NewAuthHandler(authUseCase)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware for frontend
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public auth routes (no JWT required)
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected API routes (require JWT)
	api := r.Group("/api/v1")
	{
		// Apply JWT middleware to all routes in this group
		api.Use(middleware.JWTAuthMiddleware(jwtService))

		// Club routes
		api.POST("/clubs", clubHandler.CreateClub)
		api.GET("/clubs/:club_id/members", clubHandler.GetClubMembers)
		api.POST("/clubs/:club_id/members/approve", clubHandler.ApproveMember)
		api.POST("/clubs/:club_id/members/reject", clubHandler.RejectMember)

		// Game routes
		api.POST("/games", gameHandler.CreateGame)
		api.GET("/games", gameHandler.ListGames)
		api.GET("/games/:game_id", gameHandler.GetGameDetails)
		api.GET("/games/:game_id/participants", gameHandler.GetGameParticipants)
		api.POST("/games/:game_id/buyin", gameHandler.BuyIn)
		api.POST("/games/:game_id/participants/:player_id/chips", gameHandler.SetChips)
		api.POST("/games/:game_id/finish", gameHandler.FinishGame)
		api.GET("/players/:player_id/stats", gameHandler.GetPlayerStats)

		// Leaderboard route
		api.GET("/clubs/:club_id/leaderboard", gameHandler.GetLeaderboard)
	}

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := "8080"
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

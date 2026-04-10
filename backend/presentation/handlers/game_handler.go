package handlers

import (
	"fmt"
	"net/http"

	"poker-club-backend/application/dtos"
	"poker-club-backend/application/usecases"

	"github.com/gin-gonic/gin"
)

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	gameUseCase *usecases.GameUseCase
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(gameUseCase *usecases.GameUseCase) *GameHandler {
	return &GameHandler{
		gameUseCase: gameUseCase,
	}
}

// CreateGame handles POST /games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req dtos.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.gameUseCase.CreateGame(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetGameParticipants handles GET /games/:game_id/participants
func (h *GameHandler) GetGameParticipants(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}
	var gameID int64
	if _, err := fmt.Sscanf(gameIDStr, "%d", &gameID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	participants, err := h.gameUseCase.GetGameParticipants(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game_id":      gameID,
		"participants": participants,
	})
}

// BuyIn handles POST /games/:game_id/buyin
func (h *GameHandler) BuyIn(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}
	var gameID int64
	if _, err := fmt.Sscanf(gameIDStr, "%d", &gameID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	var req dtos.BuyInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.GameID = gameID
	req.PerformedBy = c.GetInt64("user_id")

	response, err := h.gameUseCase.BuyIn(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SetChips handles POST /games/:game_id/participants/:player_id/chips
func (h *GameHandler) SetChips(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	playerIDStr := c.Param("player_id")
	if gameIDStr == "" || playerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id and player_id are required"})
		return
	}
	var gameID, playerID int64
	if _, err := fmt.Sscanf(gameIDStr, "%d", &gameID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}
	if _, err := fmt.Sscanf(playerIDStr, "%d", &playerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player_id"})
		return
	}

	var req dtos.SetChipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.GameID = gameID
	req.PlayerID = playerID
	req.PerformedBy = c.GetInt64("user_id")

	err := h.gameUseCase.SetChips(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "chips set"})
}

// FinishGame handles POST /games/:game_id/finish
func (h *GameHandler) FinishGame(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}
	var gameID int64
	if _, err := fmt.Sscanf(gameIDStr, "%d", &gameID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	var req dtos.FinishGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.GameID = gameID
	req.PerformedBy = c.GetInt64("user_id")

	response, err := h.gameUseCase.FinishGame(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPlayerStats handles GET /players/:player_id/stats
func (h *GameHandler) GetPlayerStats(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	if playerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id is required"})
		return
	}
	var playerID int64
	if _, err := fmt.Sscanf(playerIDStr, "%d", &playerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player_id"})
		return
	}

	stats, err := h.gameUseCase.GetPlayerStats(c.Request.Context(), playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ListGames handles GET /games
func (h *GameHandler) ListGames(c *gin.Context) {
	// Parse query parameters
	clubIDStr := c.Query("club_id")
	if clubIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "club_id is required"})
		return
	}
	var clubID int64
	if _, err := fmt.Sscanf(clubIDStr, "%d", &clubID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club_id"})
		return
	}

	status := c.Query("status")
	limit := 50 // default
	offset := 0

	if lim, ok := c.GetQuery("limit"); ok {
		if _, err := fmt.Sscanf(lim, "%d", &limit); err != nil || limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
	}
	if off, ok := c.GetQuery("offset"); ok {
		if _, err := fmt.Sscanf(off, "%d", &offset); err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
	}

	req := dtos.ListGamesRequest{
		ClubID: clubID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.gameUseCase.ListGames(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetGameDetails handles GET /games/:game_id
func (h *GameHandler) GetGameDetails(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}
	var gameID int64
	if _, err := fmt.Sscanf(gameIDStr, "%d", &gameID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	response, err := h.gameUseCase.GetGameDetails(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLeaderboard handles GET /clubs/:club_id/leaderboard
func (h *GameHandler) GetLeaderboard(c *gin.Context) {
	clubIDStr := c.Param("club_id")
	if clubIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "club_id is required"})
		return
	}
	var clubID int64
	if _, err := fmt.Sscanf(clubIDStr, "%d", &clubID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club_id"})
		return
	}

	metric := c.Query("metric")
	period := c.Query("period")

	if metric == "" || period == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric and period are required"})
		return
	}

	req := dtos.GetLeaderboardRequest{
		ClubID: clubID,
		Metric: metric,
		Period: period,
	}

	response, err := h.gameUseCase.GetLeaderboard(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

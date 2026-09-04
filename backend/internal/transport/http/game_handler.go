package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/service"
)

// GameHandler handles HTTP requests for game management.
// Endpoints:
//   - GET    /api/v1/clubs/{clubId}/games                           — list games in a club
//   - POST   /api/v1/clubs/{clubId}/games                           — create a game
//   - GET    /api/v1/games/{gameId}                                 — get game info
//   - PATCH  /api/v1/games/{gameId}                                 — update game parameters
//   - POST   /api/v1/games/{gameId}/start                           — start a game
//   - POST   /api/v1/games/{gameId}/finish                          — finish a game
//   - POST   /api/v1/games/{gameId}/cancel                          — cancel a game
//   - PATCH  /api/v1/games/{gameId}/banker                          — assign/change banker
//   - GET    /api/v1/games/{gameId}/participants                    — list game participants
//   - POST   /api/v1/games/{gameId}/participants                    — add player to game
//   - DELETE /api/v1/games/{gameId}/participants/{playerId}         — remove player from game
//   - POST   /api/v1/games/{gameId}/participants/{playerId}/buy-in  — register buy-in
//   - POST   /api/v1/games/{gameId}/participants/{playerId}/rebuy   — register rebuy
//   - POST   /api/v1/games/{gameId}/participants/{playerId}/chips   — set chips end
//   - GET    /api/v1/games/{gameId}/results                         — get game results
//   - POST   /api/v1/games/{gameId}/results/correct                 — correct game results
//   - GET    /api/v1/games/{gameId}/monitor                         — get game monitor data
//   - GET    /api/v1/games/{gameId}/events                          — get game events
type GameHandler struct {
	svc *service.Service
}

// NewGameHandler creates a new GameHandler.
func NewGameHandler(svc *service.Service) *GameHandler {
	return &GameHandler{svc: svc}
}

// createGameRequest represents the request body for creating a game.
type createGameRequest struct {
	GameType         string     `json:"game_type" binding:"required"`
	Currency         string     `json:"currency"`
	MoneyModel       string     `json:"money_model"`
	ChipValue        float64    `json:"chip_value"`
	BuyInAmount      float64    `json:"buy_in_amount"`
	RebuyAllowed     bool       `json:"rebuy_allowed"`
	RebuyPrice       *float64   `json:"rebuy_price,omitempty"`
	MaxRebuys        *int       `json:"max_rebuys,omitempty"`
	Duration         *int64     `json:"duration_seconds,omitempty"`
	StartTime        *time.Time `json:"start_time,omitempty"`
	MinPlayers       int        `json:"min_players"`
	MaxPlayers       int        `json:"max_players"`
	RankingPrimary   string     `json:"ranking_primary"`
	RankingSecondary *string    `json:"ranking_secondary,omitempty"`
	BankerID         int64      `json:"banker_id"`
}

// updateGameRequest represents the request body for updating a game.
type updateGameRequest struct {
	GameType         *string    `json:"game_type,omitempty"`
	Currency         *string    `json:"currency,omitempty"`
	MoneyModel       *string    `json:"money_model,omitempty"`
	ChipValue        *float64   `json:"chip_value,omitempty"`
	BuyInAmount      *float64   `json:"buy_in_amount,omitempty"`
	RebuyAllowed     *bool      `json:"rebuy_allowed,omitempty"`
	RebuyPrice       *float64   `json:"rebuy_price,omitempty"`
	MaxRebuys        *int       `json:"max_rebuys,omitempty"`
	Duration         *int64     `json:"duration_seconds,omitempty"`
	StartTime        *time.Time `json:"start_time,omitempty"`
	MinPlayers       *int       `json:"min_players,omitempty"`
	MaxPlayers       *int       `json:"max_players,omitempty"`
	RankingPrimary   *string    `json:"ranking_primary,omitempty"`
	RankingSecondary *string    `json:"ranking_secondary,omitempty"`
}

// chipsRequest represents the request body for setting chips end.
type chipsRequest struct {
	ChipsEnd float64 `json:"chips_end" binding:"required"`
}

// correctResultsRequest represents the request body for correcting game results.
type correctResultsRequest struct {
	PlayerID int64   `json:"player_id" binding:"required"`
	ChipsEnd float64 `json:"chips_end" binding:"required"`
}

// bankerRequest represents the request body for assigning a banker.
type bankerRequest struct {
	BankerID int64 `json:"banker_id" binding:"required"`
}

// addParticipantRequest represents the request body for adding a participant.
type addParticipantRequest struct {
	PlayerID int64 `json:"player_id" binding:"required"`
}

// ListGames handles GET /api/v1/clubs/{clubId}/games.
// Returns all games for a club.
func (h *GameHandler) ListGames(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	clubID, err := parseIDParam(c, "clubId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	games, err := h.svc.GetClubGames(c.Request.Context(), tgUserID, clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"games": serializeGames(games),
	})
}

// CreateGame handles POST /api/v1/clubs/{clubId}/games.
// Creates a new game in the club.
func (h *GameHandler) CreateGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	clubID, err := parseIDParam(c, "clubId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	var req createGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	game := &domain.Game{
		GameType:         req.GameType,
		Currency:         req.Currency,
		MoneyModel:       req.MoneyModel,
		ChipValue:        req.ChipValue,
		BuyInAmount:      req.BuyInAmount,
		RebuyAllowed:     req.RebuyAllowed,
		RebuyPrice:       req.RebuyPrice,
		MaxRebuys:        req.MaxRebuys,
		MinPlayers:       req.MinPlayers,
		MaxPlayers:       req.MaxPlayers,
		RankingPrimary:   req.RankingPrimary,
		RankingSecondary: req.RankingSecondary,
		BankerID:         req.BankerID,
	}

	if req.Duration != nil {
		dur := time.Duration(*req.Duration) * time.Second
		game.Duration = &dur
	}
	if req.StartTime != nil {
		game.StartTime = *req.StartTime
	}

	createdGame, err := h.svc.CreateGame(c.Request.Context(), tgUserID, clubID, game)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, serializeGame(createdGame))
}

// GetGame handles GET /api/v1/games/{gameId}.
// Returns information about a specific game.
func (h *GameHandler) GetGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeGame(game))
}

// UpdateGame handles PATCH /api/v1/games/{gameId}.
// Updates game parameters. Only allowed in planned status.
func (h *GameHandler) UpdateGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	var req updateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	// Apply updates.
	if req.GameType != nil {
		game.GameType = *req.GameType
	}
	if req.Currency != nil {
		game.Currency = *req.Currency
	}
	if req.MoneyModel != nil {
		game.MoneyModel = *req.MoneyModel
	}
	if req.ChipValue != nil {
		game.ChipValue = *req.ChipValue
	}
	if req.BuyInAmount != nil {
		game.BuyInAmount = *req.BuyInAmount
	}
	if req.RebuyAllowed != nil {
		game.RebuyAllowed = *req.RebuyAllowed
	}
	if req.RebuyPrice != nil {
		game.RebuyPrice = req.RebuyPrice
	}
	if req.MaxRebuys != nil {
		game.MaxRebuys = req.MaxRebuys
	}
	if req.Duration != nil {
		dur := time.Duration(*req.Duration) * time.Second
		game.Duration = &dur
	}
	if req.StartTime != nil {
		game.StartTime = *req.StartTime
	}
	if req.MinPlayers != nil {
		game.MinPlayers = *req.MinPlayers
	}
	if req.MaxPlayers != nil {
		game.MaxPlayers = *req.MaxPlayers
	}
	if req.RankingPrimary != nil {
		game.RankingPrimary = *req.RankingPrimary
	}
	if req.RankingSecondary != nil {
		game.RankingSecondary = req.RankingSecondary
	}

	updatedGame, err := h.svc.UpdateGame(c.Request.Context(), tgUserID, game.ClubID, gameID, game)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeGame(updatedGame))
}

// StartGame handles POST /api/v1/games/{gameId}/start.
// Starts a game (transitions from planned to active).
func (h *GameHandler) StartGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	updatedGame, err := h.svc.StartGame(c.Request.Context(), tgUserID, game.ClubID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeGame(updatedGame))
}

// FinishGame handles POST /api/v1/games/{gameId}/finish.
// Finishes a game (transitions from active to finished).
func (h *GameHandler) FinishGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.FinishGame(c.Request.Context(), tgUserID, game.ClubID, gameID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "game finished successfully",
	})
}

// CancelGame handles POST /api/v1/games/{gameId}/cancel.
// Cancels a game (transitions to cancelled status).
func (h *GameHandler) CancelGame(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.CancelGame(c.Request.Context(), tgUserID, game.ClubID, gameID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "game cancelled successfully",
	})
}

// UpdateBanker handles PATCH /api/v1/games/{gameId}/banker.
// Assigns or changes the banker for a game.
func (h *GameHandler) UpdateBanker(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	var req bankerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	updatedGame, err := h.svc.ChangeBanker(c.Request.Context(), tgUserID, game.ClubID, gameID, req.BankerID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeGame(updatedGame))
}

// ListGameParticipants handles GET /api/v1/games/{gameId}/participants.
// Returns all participants of a game.
func (h *GameHandler) ListGameParticipants(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	participants, err := h.svc.GetGameParticipants(c.Request.Context(), tgUserID, game.ClubID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participants": serializeGameParticipants(participants),
	})
}

// AddGameParticipant handles POST /api/v1/games/{gameId}/participants.
// Adds a player to a game.
func (h *GameHandler) AddGameParticipant(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	var req addParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.AddPlayerToGame(c.Request.Context(), tgUserID, game.ClubID, gameID, req.PlayerID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "player added to game",
	})
}

// RemoveGameParticipant handles DELETE /api/v1/games/{gameId}/participants/{playerId}.
// Removes a player from a game.
func (h *GameHandler) RemoveGameParticipant(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	player, err := h.svc.RemoveGameParticipant(c.Request.Context(), tgUserID, game.ClubID, gameID, playerID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player":  serializePlayer(player),
		"message": "player removed from game",
	})
}

// RegisterBuyIn handles POST /api/v1/games/{gameId}/participants/{playerId}/buy-in.
// Registers a buy-in for a game participant.
func (h *GameHandler) RegisterBuyIn(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.RegisterBuyIn(c.Request.Context(), tgUserID, game.ClubID, gameID, playerID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "buy-in registered",
	})
}

// RegisterRebuy handles POST /api/v1/games/{gameId}/participants/{playerId}/rebuy.
// Registers a rebuy for a game participant.
func (h *GameHandler) RegisterRebuy(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.RegisterRebuy(c.Request.Context(), tgUserID, game.ClubID, gameID, playerID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "rebuy registered",
	})
}

// SetChipsEnd handles POST /api/v1/games/{gameId}/participants/{playerId}/chips.
// Sets the final chips count for a game participant.
func (h *GameHandler) SetChipsEnd(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	var req chipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.SetChipsEnd(c.Request.Context(), tgUserID, game.ClubID, gameID, playerID, req.ChipsEnd); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "chips end set",
	})
}

// GetGameResults handles GET /api/v1/games/{gameId}/results.
// Returns the results of a finished game.
func (h *GameHandler) GetGameResults(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	participants, err := h.svc.GetGameParticipants(c.Request.Context(), tgUserID, game.ClubID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	// Build results from participants.
	results := make([]gin.H, len(participants))
	for i, p := range participants {
		result := gin.H{
			"player_id":     p.PlayerID,
			"buy_in_count":  p.BuyInCount,
			"rebuy_count":   p.RebuyCount,
			"chips_end":     p.ChipsEnd,
			"payout_amount": p.PayoutAmount,
			"place":         p.Place,
			"status":        p.Status,
		}
		results[i] = result
	}

	c.JSON(http.StatusOK, gin.H{
		"game":       serializeGame(game),
		"results":    results,
	})
}

// CorrectGameResults handles POST /api/v1/games/{gameId}/results/correct.
// Corrects game results after the game has finished.
func (h *GameHandler) CorrectGameResults(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	var req correctResultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	if err := h.svc.AdjustGameResults(c.Request.Context(), tgUserID, game.ClubID, gameID, req.PlayerID, req.ChipsEnd); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "game results corrected",
	})
}

// GetGameMonitor handles GET /api/v1/games/{gameId}/monitor.
// Returns game monitor data (game state and participants).
func (h *GameHandler) GetGameMonitor(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	g, participants, err := h.svc.GetGameMonitor(c.Request.Context(), tgUserID, game.ClubID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game":       serializeGame(g),
		"participants": serializeGameParticipants(participants),
	})
}

// GetGameEvents handles GET /api/v1/games/{gameId}/events.
// Returns the event log for a game.
func (h *GameHandler) GetGameEvents(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	gameID, err := parseIDParam(c, "gameId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid game ID")
		return
	}

	// Get the game to resolve clubID.
	game, err := h.svc.GetGameByID(c.Request.Context(), tgUserID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	events, err := h.svc.GetGameEvents(c.Request.Context(), tgUserID, game.ClubID, gameID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": serializeEvents(events),
	})
}

// serializeGame converts a domain.Game to a JSON-serializable map.
func serializeGame(g *domain.Game) gin.H {
	result := gin.H{
		"id":                 g.ID,
		"club_id":            g.ClubID,
		"banker_id":          g.BankerID,
		"game_type":          g.GameType,
		"currency":           g.Currency,
		"money_model":        g.MoneyModel,
		"chip_value":         g.ChipValue,
		"buy_in_amount":      g.BuyInAmount,
		"rebuy_allowed":      g.RebuyAllowed,
		"rebuy_price":        g.RebuyPrice,
		"max_rebuys":         g.MaxRebuys,
		"duration_seconds":   nil,
		"start_time":         g.StartTime,
		"end_time":           g.EndTime,
		"status":             g.Status,
		"min_players":        g.MinPlayers,
		"max_players":        g.MaxPlayers,
		"ranking_primary":    g.RankingPrimary,
		"ranking_secondary":  g.RankingSecondary,
		"created_at":         g.CreatedAt,
		"updated_at":         g.UpdatedAt,
		"timer_paused_at":    g.TimerPausedAt,
		"timer_paused_duration": nil,
		"timer_notified":     g.TimerNotified,
	}
	if g.Duration != nil {
		result["duration_seconds"] = int64(g.Duration.Seconds())
	}
	if g.TimerPausedDuration != nil {
		result["timer_paused_duration"] = int64(g.TimerPausedDuration.Seconds())
	}
	return result
}

// serializeGames converts a list of domain.Game to a JSON-serializable slice.
func serializeGames(games []*domain.Game) []gin.H {
	result := make([]gin.H, len(games))
	for i, g := range games {
		result[i] = serializeGame(g)
	}
	return result
}

// serializeGameParticipant converts a domain.GameParticipantWithPlayer to a JSON-serializable map.
func serializeGameParticipant(p *domain.GameParticipantWithPlayer) gin.H {
	result := gin.H{
		"id":            p.ID,
		"game_id":       p.GameID,
		"player_id":     p.PlayerID,
		"buy_in_count":  p.BuyInCount,
		"rebuy_count":   p.RebuyCount,
		"chips_end":     p.ChipsEnd,
		"payout_amount": p.PayoutAmount,
		"place":         p.Place,
		"status":        p.Status,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
		"player":        serializePlayer(&p.Player),
	}
	return result
}

// serializeGameParticipants converts a list of domain.GameParticipantWithPlayer to a JSON-serializable slice.
func serializeGameParticipants(participants []*domain.GameParticipantWithPlayer) []gin.H {
	result := make([]gin.H, len(participants))
	for i, p := range participants {
		result[i] = serializeGameParticipant(p)
	}
	return result
}

// serializeEvent converts a domain.Event to a JSON-serializable map.
func serializeEvent(e *domain.Event) gin.H {
	result := gin.H{
		"id":         e.ID,
		"game_id":    e.GameID,
		"player_id":  e.PlayerID,
		"type":       e.Type,
		"old_value":  e.OldValue,
		"new_value":  e.NewValue,
		"metadata":   e.Metadata,
		"created_at": e.CreatedAt,
		"created_by": e.CreatedBy,
	}
	return result
}

// serializeEvents converts a list of domain.Event to a JSON-serializable slice.
func serializeEvents(events []*domain.Event) []gin.H {
	result := make([]gin.H, len(events))
	for i, e := range events {
		result[i] = serializeEvent(e)
	}
	return result
}

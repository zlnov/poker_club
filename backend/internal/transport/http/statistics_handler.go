package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/service"
)

// StatisticsHandler handles HTTP requests for statistics.
// Endpoints:
//   - GET /api/v1/clubs/{clubId}/statistics     — club statistics
//   - GET /api/v1/players/{playerId}/statistics — player statistics (requires clubId query param)
type StatisticsHandler struct {
	svc *service.Service
}

// NewStatisticsHandler creates a new StatisticsHandler.
func NewStatisticsHandler(svc *service.Service) *StatisticsHandler {
	return &StatisticsHandler{svc: svc}
}

// GetClubStatistics handles GET /api/v1/clubs/{clubId}/statistics.
// Returns aggregate statistics for a club.
func (h *StatisticsHandler) GetClubStatistics(c *gin.Context) {
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

	stats, err := h.svc.GetClubStatistics(c.Request.Context(), tgUserID, clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_members":         stats.TotalMembers,
		"total_games":           stats.TotalGames,
		"cash_games":            stats.CashGames,
		"tournament_games":      stats.TournamentGames,
		"total_buy_in_amount":   stats.TotalBuyInAmount,
		"total_rebuy_amount":    stats.TotalRebuyAmount,
		"total_bank":            stats.TotalBank,
		"average_game_duration": int64(stats.AverageGameDuration.Seconds()),
	})
}

// GetPlayerStatistics handles GET /api/v1/players/{playerId}/statistics.
// Returns aggregate statistics for a player in a club.
// Requires a clubId query parameter to identify the club context.
// Optional game_type query parameter filters statistics by game type (cash/tournament).
func (h *StatisticsHandler) GetPlayerStatistics(c *gin.Context) {
	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	// The clubId is required as a query parameter.
	clubIDStr := c.Query("clubId")
	if clubIDStr == "" {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "clubId query parameter is required")
		return
	}
	clubID, err := strconv.ParseInt(clubIDStr, 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	// Optional game_type query parameter.
	gameType := c.Query("game_type")
	if gameType == "" {
		gameType = "cash"
	}

	// Verify the requesting user is a member of the club.
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	// Check membership via GetClubMembers (which performs permission check).
	members, err := h.svc.GetClubMembers(c.Request.Context(), tgUserID, clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	// Verify the requesting user is a member.
	found := false
	for _, m := range members {
		if m.PlayerID == playerID || m.Player.TgUserID != nil && *m.Player.TgUserID == tgUserID {
			found = true
			break
		}
	}
	if !found {
		writeError(c, http.StatusForbidden, "CLUB_ACCESS_DENIED", "access denied: user is not a member of this club")
		return
	}

	// Use the game_type-aware service method.
	stats, err := h.svc.GetPlayerStatisticsWithGameType(c.Request.Context(), playerID, clubID, gameType)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_games":           stats.TotalGames,
		"total_buy_in_amount":   stats.TotalBuyInAmount,
		"total_rebuy_amount":    stats.TotalRebuyAmount,
		"total_rebuy_count":     stats.TotalRebuysCount,
		"total_invested":        stats.TotalInvested,
		"total_chips":           stats.TotalChips,
		"total_profit":          stats.TotalProfit,
		"biggest_win":           stats.BiggestWin,
		"biggest_loss":          stats.BiggestLoss,
		"games_won":             stats.GamesWon,
		"podiums":               stats.Podiums,
		"roi":                   stats.ROI,
		"itm":                   stats.ITM,
		"total_buy_in_count":    stats.TotalBuyInCount,
		"winrate":               stats.Winrate,
		"avg_place":             stats.AvgPlace,
		"games_in_profit":       stats.GamesInProfit,
	})
}

// GetClubMemberStatistics handles GET /api/v1/clubs/{clubId}/statistics/members.
// Returns aggregated statistics for each club member filtered by game_type.
func (h *StatisticsHandler) GetClubMemberStatistics(c *gin.Context) {
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

	// Optional game_type query parameter.
	gameType := c.Query("game_type")
	if gameType == "" {
		gameType = "cash"
	}

	members, err := h.svc.GetClubMemberStatistics(c.Request.Context(), tgUserID, clubID, gameType)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	responseMembers := make([]gin.H, len(members))
	for i, m := range members {
		responseMembers[i] = gin.H{
			"player_id":      m.PlayerID,
			"player_name":    m.PlayerName,
			"games":          m.Games,
			"total_invested": m.TotalInvested,
			"profit":         m.Profit,
			"roi":            m.ROI,
			"winrate":        m.Winrate,
			"avg_place":      m.AvgPlace,
			"games_won":      m.GamesWon,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"members": responseMembers,
	})
}

// GetPlayerGameHistory handles GET /api/v1/players/{playerId}/history.
// Returns the game history for a player filtered by game_type.
func (h *StatisticsHandler) GetPlayerGameHistory(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	clubIDStr := c.Query("clubId")
	if clubIDStr == "" {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "clubId query parameter is required")
		return
	}
	clubID, err := strconv.ParseInt(clubIDStr, 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	// Optional game_type query parameter.
	gameType := c.Query("game_type")
	if gameType == "" {
		gameType = "cash"
	}

	history, err := h.svc.GetPlayerGameHistory(c.Request.Context(), tgUserID, clubID, playerID, gameType)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	responseHistory := make([]gin.H, len(history))
	for i, h := range history {
		responseHistory[i] = gin.H{
			"game_id":        h.GameID,
			"game_name":      h.GameName,
			"player_id":      h.PlayerID,
			"player_name":    h.PlayerName,
			"place":          h.Place,
			"buy_in_count":   h.BuyInCount,
			"rebuy_count":    h.RebuyCount,
			"buy_in_amount":  h.BuyInAmount,
			"rebuy_amount":   h.RebuyAmount,
			"total_invested": h.TotalInvested,
			"chips_end":      h.ChipsEnd,
			"payout_amount":  h.PayoutAmount,
			"profit":         h.Profit,
			"roi":            h.ROI,
			"status":         h.Status,
			"game_type":      h.GameType,
			"start_time":     h.StartTime,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"history": responseHistory,
	})
}

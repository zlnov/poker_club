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

	stats, err := h.svc.GetPlayerStatisticsByID(c.Request.Context(), playerID, clubID)
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
	})
}

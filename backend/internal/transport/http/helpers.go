package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// getPlayerIDFromContext extracts the authenticated player's database ID from the Gin context.
// The ID is set by AuthMiddleware from the JWT claims.
func getPlayerIDFromContext(c *gin.Context) (int64, bool) {
	user, ok := GetAuthenticatedUser(c)
	if !ok {
		return 0, false
	}
	return user.PlayerID, true
}

// getTgUserIDFromContext extracts the authenticated user's Telegram user ID from the Gin context.
// Returns the TgUserID and true if available, or 0 and false if not set.
func getTgUserIDFromContext(c *gin.Context) (int64, bool) {
	user, ok := GetAuthenticatedUser(c)
	if !ok || user.TgUserID == nil {
		return 0, false
	}
	return *user.TgUserID, true
}

// parseIDParam extracts and parses an integer ID parameter from the URL path.
// Returns the parsed ID and an error if parsing fails.
func parseIDParam(c *gin.Context, name string) (int64, error) {
	val := c.Param(name)
	return strconv.ParseInt(val, 10, 64)
}

// writeError writes a standard error response with the given code and HTTP status.
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse(code, message))
}

// writeServiceError writes an error response based on a service error.
// It maps common service error patterns to appropriate HTTP status codes and error codes.
func writeServiceError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	msg := err.Error()

	// Map known error messages to appropriate status codes and error codes.
	switch {
	case strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "не хватает прав") ||
		strings.Contains(msg, "insufficient permissions") ||
		strings.Contains(msg, "not a member") ||
		strings.Contains(msg, "не является участником"):
		writeError(c, http.StatusForbidden, "CLUB_ACCESS_DENIED", msg)
	case strings.Contains(msg, "not found") ||
		strings.Contains(msg, "не найден") ||
		strings.Contains(msg, "не найдена"):
		writeError(c, http.StatusNotFound, "CLUB_NOT_FOUND", msg)
	case strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "уже является") ||
		strings.Contains(msg, "уже обработан"):
		writeError(c, http.StatusConflict, "MEMBER_ALREADY_EXISTS", msg)
	case strings.Contains(msg, "cannot") ||
		strings.Contains(msg, "нельзя") ||
		strings.Contains(msg, "не доступно") ||
		strings.Contains(msg, "доступно только") ||
		strings.Contains(msg, "уже есть"):
		writeError(c, http.StatusConflict, "CLUB_INVALID_STATE", msg)
	case strings.Contains(msg, "not allowed") ||
		strings.Contains(msg, "не разрешен"):
		writeError(c, http.StatusForbidden, "CLUB_ACCESS_DENIED", msg)
	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", msg)
	}
}

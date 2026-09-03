package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/auth"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	uc *auth.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc *auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// loginRequest represents the request body for login.
type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// telegramAuthRequest represents the request body for Telegram authentication.
type telegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

// refreshRequest represents the request body for token refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// logoutRequest represents the request body for logout.
type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Login handles POST /api/v1/auth/login
// Authenticates a user with login + password and returns JWT tokens.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	tokens, err := h.uc.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
	})
}

// TelegramAuth handles POST /api/v1/auth/telegram
// Validates Telegram initData and returns JWT tokens.
func (h *AuthHandler) TelegramAuth(c *gin.Context) {
	var req telegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	tokens, err := h.uc.AuthenticateTelegram(c.Request.Context(), req.InitData)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
	})
}

// Refresh handles POST /api/v1/auth/refresh
// Refreshes the access token using a valid refresh token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	tokens, err := h.uc.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
	})
}

// Logout handles POST /api/v1/auth/logout
// Revokes the refresh token and invalidates the session.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	// Refresh token is optional — if not provided, we just clear the session.
	_ = c.ShouldBindJSON(&req)

	_ = h.uc.Logout(c.Request.Context(), req.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Me handles GET /api/v1/me
// Returns the current authenticated player.
func (h *AuthHandler) Me(c *gin.Context) {
	user, exists := GetAuthenticatedUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("AUTHENTICATION_REQUIRED", "authentication required"))
		return
	}

	player, err := h.uc.GetCurrentUser(c.Request.Context(), user.PlayerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_SERVER_ERROR", "failed to get user"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           player.ID,
		"first_name":   player.FirstName,
		"last_name":    player.LastName,
		"nickname":     player.Nickname,
		"tg_user_id":   player.TgUserID,
		"created_at":   player.CreatedAt,
		"updated_at":   player.UpdatedAt,
	})
}

// errorResponse creates a standard error response.
func errorResponse(code, message string) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

// writeAuthError writes an authentication error response based on the error type.
func writeAuthError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *auth.AuthError:
		c.JSON(http.StatusUnauthorized, errorResponse(e.Code, e.Msg))
	case *auth.TelegramInitDataError:
		c.JSON(http.StatusUnauthorized, errorResponse(e.Code, e.Msg))
	default:
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_SERVER_ERROR", "internal server error"))
	}
}

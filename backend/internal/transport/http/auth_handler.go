package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/auth"
	"poker-club/backend/internal/domain"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	uc *auth.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc *auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type telegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

type telegramWebAuthRequest struct {
	IDToken        string `json:"id_token" binding:"required"`
	ChallengeToken string `json:"challenge_token" binding:"required"`
}

type telegramWebRegisterRequest struct {
	RegistrationToken    string `json:"registration_token" binding:"required"`
	Email                string `json:"email" binding:"required"`
	Password             string `json:"password" binding:"required"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required"`
	Nickname             string `json:"nickname" binding:"required"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	tokens, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	writeTokenJSON(c, h.uc, tokens)
}

// TelegramAuth handles POST /api/v1/auth/telegram (Mini App initData).
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

	writeTokenJSON(c, h.uc, tokens)
}

// TelegramWebChallenge handles POST /api/v1/auth/telegram/web/challenge
func (h *AuthHandler) TelegramWebChallenge(c *gin.Context) {
	result, err := h.uc.CreateTelegramWebChallenge()
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"challenge_token": result.ChallengeToken,
		"nonce":           result.Nonce,
		"expires_in":      result.ExpiresIn,
	})
}

// TelegramWebAuth handles POST /api/v1/auth/telegram/web
func (h *AuthHandler) TelegramWebAuth(c *gin.Context) {
	var req telegramWebAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	result, err := h.uc.AuthenticateTelegramWeb(c.Request.Context(), req.IDToken, req.ChallengeToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	if result.RegistrationRequired {
		c.JSON(http.StatusOK, gin.H{
			"registration_required": true,
			"registration_token":    result.RegistrationToken,
		})
		return
	}

	writeTokenJSON(c, h.uc, result.Tokens)
}

// TelegramWebRegister handles POST /api/v1/auth/telegram/web/register
func (h *AuthHandler) TelegramWebRegister(c *gin.Context) {
	var req telegramWebRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	tokens, err := h.uc.CompleteTelegramRegistration(c.Request.Context(), auth.CompleteTelegramRegistrationRequest{
		RegistrationToken:    req.RegistrationToken,
		Email:                req.Email,
		Password:             req.Password,
		PasswordConfirmation: req.PasswordConfirmation,
		Nickname:             req.Nickname,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}

	writeTokenJSON(c, h.uc, tokens)
}

// Refresh handles POST /api/v1/auth/refresh
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

	writeTokenJSON(c, h.uc, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	_ = c.ShouldBindJSON(&req)

	_ = h.uc.Logout(c.Request.Context(), req.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Me handles GET /api/v1/me
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

	c.JSON(http.StatusOK, serializeMe(player))
}

type updateProfileRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Nickname    *string `json:"nickname"`
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
}

// UpdateMe handles PATCH /api/v1/me
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	user, exists := GetAuthenticatedUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("AUTHENTICATION_REQUIRED", "authentication required"))
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	player, err := h.uc.GetCurrentUser(c.Request.Context(), user.PlayerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_SERVER_ERROR", "failed to get user"))
		return
	}

	update := auth.UpdateProfileRequest{
		FirstName: player.FirstName,
		LastName:  player.LastName,
		Nickname:  player.NicknameOrEmpty(),
		Email:     req.Email,
		PhoneNumber: req.PhoneNumber,
	}
	if req.FirstName != nil {
		update.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		update.LastName = *req.LastName
	}
	if req.Nickname != nil {
		update.Nickname = *req.Nickname
	}

	updated, err := h.uc.UpdateProfile(c.Request.Context(), user.PlayerID, update)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeMe(updated))
}

type changePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	NewPassword          string `json:"new_password" binding:"required"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required"`
}

// ChangePassword handles POST /api/v1/me/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, exists := GetAuthenticatedUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse("AUTHENTICATION_REQUIRED", "authentication required"))
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "invalid request"))
		return
	}

	if err := h.uc.ChangePassword(c.Request.Context(), user.PlayerID, auth.ChangePasswordRequest{
		CurrentPassword:      req.CurrentPassword,
		NewPassword:          req.NewPassword,
		PasswordConfirmation: req.PasswordConfirmation,
	}); err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func serializeMe(player *domain.Player) gin.H {
	return gin.H{
		"id":           player.ID,
		"first_name":   player.FirstName,
		"last_name":    player.LastName,
		"nickname":     player.Nickname,
		"tg_user_name": player.TgUserName,
		"email":        player.Email,
		"phone_number": player.PhoneNumber,
		"tg_user_id":   player.TgUserID,
		"has_password": player.HasPassword(),
		"created_at":   player.CreatedAt,
		"updated_at":   player.UpdatedAt,
	}
}

func writeTokenJSON(c *gin.Context, uc *auth.AuthUseCase, tokens *domain.AuthTokens) {
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(uc.AccessTokenTTL().Seconds()),
	})
}

func errorResponse(code, message string) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

func writeAuthError(c *gin.Context, err error) {
	var authErr *auth.AuthError
	if errors.As(err, &authErr) {
		c.JSON(authErr.HTTPStatusHint(), errorResponse(authErr.Code, authErr.Msg))
		return
	}
	var tgErr *auth.TelegramInitDataError
	if errors.As(err, &tgErr) {
		c.JSON(http.StatusUnauthorized, errorResponse(tgErr.Code, tgErr.Msg))
		return
	}
	c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_SERVER_ERROR", "internal server error"))
}

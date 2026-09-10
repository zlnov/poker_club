package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"poker-club/backend/internal/domain"
)

// AuthUseCase provides authentication business logic.
type AuthUseCase struct {
	repos     *domain.Repositories
	jwt       *JWTManager
	botToken  string
	log       *slog.Logger
	maxAge    time.Duration
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(repos *domain.Repositories, jwt *JWTManager, botToken string, log *slog.Logger) *AuthUseCase {
	return &AuthUseCase{
		repos:    repos,
		jwt:      jwt,
		botToken: botToken,
		log:      log,
		maxAge:   24 * time.Hour, // Telegram initData max age
	}
}

// Login authenticates a user by login (nickname) and password.
// Returns INVALID_CREDENTIALS if the credentials are invalid.
func (uc *AuthUseCase) Login(ctx context.Context, login, password string) (*domain.AuthTokens, error) {
	player, err := uc.repos.Players.GetByNickname(ctx, login)
	if err != nil {
		// Don't reveal whether the user exists.
		uc.log.Warn("login failed: player not found", "login", login)
		return nil, ErrInvalidCredentials
	}

	// Compare password hash.
	if err := bcrypt.CompareHashAndPassword([]byte(player.Password), []byte(password)); err != nil {
		uc.log.Warn("login failed: invalid password", "login", login)
		return nil, ErrInvalidCredentials
	}

	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}

	return uc.issueTokens(ctx, authUser)
}

// AuthenticateTelegram validates Telegram initData and returns auth tokens.
// If the player doesn't exist, it creates a basic player record.
// Does NOT create Club Membership.
func (uc *AuthUseCase) AuthenticateTelegram(ctx context.Context, initData string) (*domain.AuthTokens, error) {
	// Validate the Telegram initData.
	parsed, err := ValidateTelegramInitData(initData, uc.botToken, uc.maxAge)
	if err != nil {
		uc.log.Warn("telegram auth validation failed",
			"error", err.Error(),
			"botToken_len", len(uc.botToken),
			"initData_len", len(initData),
		)
		if err == ErrExpiredInitData {
			return nil, ErrTelegramInitDataExpired
		}
		return nil, ErrInvalidTelegramInitData
	}

	// Find or create the player.
	player, err := uc.repos.Players.GetByTgUserID(ctx, parsed.UserID)
	if err != nil {
		// Player doesn't exist — create a basic player record.
		player = &domain.Player{
			FirstName: parsed.FirstName,
			LastName:  parsed.LastName,
			Nickname:  parsed.Username,
			Password:  "",
			TgUserID:  &parsed.UserID,
		}
		playerID, err := uc.repos.Players.Create(ctx, player)
		if err != nil {
			return nil, fmt.Errorf("failed to create player from telegram: %w", err)
		}
		player.ID = playerID
		uc.log.Info("player created from telegram initData",
			"player_id", player.ID,
			"tg_user_id", parsed.UserID,
		)
	} else {
		// Update last seen.
		_ = uc.repos.Players.UpdateLastSeen(ctx, player.ID)
	}

	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}

	return uc.issueTokens(ctx, authUser)
}

// RefreshAccessToken issues a new access token using a valid refresh token.
func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	// Hash the provided token and look it up.
	tokenHash := uc.jwt.HashToken(refreshToken)
	stored, err := uc.repos.RefreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if the token is revoked.
	if stored.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check expiration.
	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	// Get the player.
	player, err := uc.repos.Players.GetByID(ctx, stored.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}

	return uc.issueTokens(ctx, authUser)
}

// Logout revokes the refresh token and invalidates the session.
func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil // Nothing to revoke.
	}

	tokenHash := uc.jwt.HashToken(refreshToken)
	_ = uc.repos.RefreshTokens.Revoke(ctx, tokenHash)
	return nil
}

// LogoutAll revokes all refresh tokens for a player.
func (uc *AuthUseCase) LogoutAll(ctx context.Context, playerID int64) error {
	return uc.repos.RefreshTokens.RevokeByPlayer(ctx, playerID)
}

// GetCurrentUser returns the current authenticated player.
func (uc *AuthUseCase) GetCurrentUser(ctx context.Context, playerID int64) (*domain.Player, error) {
	return uc.repos.Players.GetByID(ctx, playerID)
}

// issueTokens generates a new access token and refresh token pair.
func (uc *AuthUseCase) issueTokens(ctx context.Context, user *domain.AuthenticatedUser) (*domain.AuthTokens, error) {
	// Generate access token.
	accessToken, err := uc.jwt.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token.
	refreshToken, err := uc.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store the hashed refresh token.
	tokenHash := uc.jwt.HashToken(refreshToken)
	expiresAt := time.Now().Add(uc.jwt.RefreshTokenTTL())
	rt := &domain.RefreshToken{
		PlayerID:  user.PlayerID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if _, err := uc.repos.RefreshTokens.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Auth errors
var (
	ErrInvalidCredentials     = &AuthError{Code: "INVALID_CREDENTIALS", Msg: "invalid credentials"}
	ErrInvalidTelegramInitData = &TelegramInitDataError{Code: "INVALID_TELEGRAM_INIT_DATA", Msg: "invalid telegram init data"}
	ErrTelegramInitDataExpired = &TelegramInitDataError{Code: "TELEGRAM_INIT_DATA_EXPIRED", Msg: "telegram init data expired"}
	ErrInvalidRefreshToken     = &AuthError{Code: "REFRESH_TOKEN_INVALID", Msg: "invalid refresh token"}
	ErrRefreshTokenExpired     = &AuthError{Code: "REFRESH_TOKEN_EXPIRED", Msg: "refresh token expired"}
	ErrAuthenticationRequired  = &AuthError{Code: "AUTHENTICATION_REQUIRED", Msg: "authentication required"}
	ErrJWTExpired              = &AuthError{Code: "JWT_EXPIRED", Msg: "jwt expired"}
	ErrJWTInvalid              = &AuthError{Code: "JWT_INVALID", Msg: "jwt invalid"}
)

// AuthError represents an authentication error with a stable error code.
type AuthError struct {
	Code string
	Msg  string
}

func (e *AuthError) Error() string {
	return e.Msg
}

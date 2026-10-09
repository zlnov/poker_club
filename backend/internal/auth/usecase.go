package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"poker-club/backend/internal/domain"
)

const (
	minPasswordLength = 12
	minNicknameLength = 2
	maxNicknameLength = 32
	bcryptCost        = bcrypt.DefaultCost
)

// AuthUseCase provides authentication business logic.
type AuthUseCase struct {
	repos     *domain.Repositories
	jwt       *JWTManager
	botToken  string
	log       *slog.Logger
	maxAge    time.Duration
	oidc      *TelegramOIDCValidator
	challenge *ChallengeManager
}

// NewAuthUseCase creates a new AuthUseCase.
// oidc and challenge may be nil when Telegram Web Login is not configured;
// Mini App initData auth still works via botToken.
func NewAuthUseCase(
	repos *domain.Repositories,
	jwt *JWTManager,
	botToken string,
	log *slog.Logger,
	oidc *TelegramOIDCValidator,
	challenge *ChallengeManager,
) *AuthUseCase {
	return &AuthUseCase{
		repos:     repos,
		jwt:       jwt,
		botToken:  botToken,
		log:       log,
		maxAge:    24 * time.Hour,
		oidc:      oidc,
		challenge: challenge,
	}
}

// Login authenticates a user by email and password.
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.AuthTokens, error) {
	player, err := uc.repos.Players.GetByEmail(ctx, email)
	if err != nil {
		uc.log.Warn("login failed: player not found or lookup error", "email", email)
		return nil, ErrInvalidCredentials
	}

	if !player.HasPassword() {
		uc.log.Warn("login failed: password not set", "email", email)
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*player.Password), []byte(password)); err != nil {
		uc.log.Warn("login failed: invalid password", "email", email)
		return nil, ErrInvalidCredentials
	}

	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}

	return uc.issueTokens(ctx, authUser)
}

// AuthenticateTelegram validates Telegram Mini App initData and returns auth tokens.
// If the player doesn't exist, it creates a basic player record (Mini App flow).
// Web OIDC must NOT use this method.
func (uc *AuthUseCase) AuthenticateTelegram(ctx context.Context, initData string) (*domain.AuthTokens, error) {
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

	tgUserName := optionalNonEmpty(parsed.Username)

	player, err := uc.repos.Players.GetByTgUserID(ctx, parsed.UserID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("failed to lookup player by tg_user_id: %w", err)
		}

		player = &domain.Player{
			FirstName:  parsed.FirstName,
			LastName:   parsed.LastName,
			Nickname:   nil,
			TgUserName: tgUserName,
			Password:   nil,
			TgUserID:   &parsed.UserID,
		}
		playerID, createErr := uc.repos.Players.Create(ctx, player)
		if createErr != nil {
			if errors.Is(createErr, domain.ErrConflict) {
				player, err = uc.repos.Players.GetByTgUserID(ctx, parsed.UserID)
				if err != nil {
					return nil, fmt.Errorf("failed to get player after conflict: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to create player from telegram: %w", createErr)
			}
		} else {
			player.ID = playerID
			uc.log.Info("player created from telegram initData",
				"player_id", player.ID,
				"tg_user_id", parsed.UserID,
			)
		}
	} else {
		_ = uc.repos.Players.UpdateLastSeen(ctx, player.ID)
		if syncErr := uc.syncTgUserName(ctx, player, tgUserName); syncErr != nil {
			uc.log.Warn("failed to sync tg_user_name", "player_id", player.ID, "error", syncErr)
		}
	}

	authUser := &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	}

	return uc.issueTokens(ctx, authUser)
}

// CreateTelegramWebChallenge issues a nonce bound in a signed challenge token.
func (uc *AuthUseCase) CreateTelegramWebChallenge() (*ChallengeResult, error) {
	if uc.challenge == nil {
		return nil, ErrTelegramOIDCInvalid
	}
	return uc.challenge.CreateChallenge()
}

// TelegramWebAuthResult is the outcome of Web Telegram OIDC authentication.
type TelegramWebAuthResult struct {
	Tokens               *domain.AuthTokens
	RegistrationRequired bool
	RegistrationToken    string
}

// AuthenticateTelegramWeb validates Telegram OIDC id_token for Web Login.
// Existing player → tokens. Missing player → registration_required (no INSERT).
func (uc *AuthUseCase) AuthenticateTelegramWeb(ctx context.Context, idToken, challengeToken string) (*TelegramWebAuthResult, error) {
	if uc.oidc == nil || uc.challenge == nil {
		return nil, ErrTelegramOIDCInvalid
	}

	nonce, err := uc.challenge.VerifyChallenge(challengeToken)
	if err != nil {
		return nil, err
	}

	identity, err := uc.oidc.ValidateIDToken(ctx, idToken, nonce)
	if err != nil {
		uc.log.Warn("telegram web oidc validation failed", "error", err.Error())
		return nil, err
	}

	tgUserName := optionalNonEmpty(identity.Username)

	player, err := uc.repos.Players.GetByTgUserID(ctx, identity.UserID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("failed to lookup player by tg_user_id: %w", err)
		}

		regToken, tokenErr := uc.challenge.CreateRegistrationToken(identity)
		if tokenErr != nil {
			return nil, fmt.Errorf("failed to create registration token: %w", tokenErr)
		}
		return &TelegramWebAuthResult{
			RegistrationRequired: true,
			RegistrationToken:    regToken,
		}, nil
	}

	_ = uc.repos.Players.UpdateLastSeen(ctx, player.ID)
	if syncErr := uc.syncTgUserName(ctx, player, tgUserName); syncErr != nil {
		uc.log.Warn("failed to sync tg_user_name", "player_id", player.ID, "error", syncErr)
	}

	tokens, err := uc.issueTokens(ctx, &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	})
	if err != nil {
		return nil, err
	}

	return &TelegramWebAuthResult{Tokens: tokens}, nil
}

// CompleteTelegramRegistrationRequest is the web registration payload.
type CompleteTelegramRegistrationRequest struct {
	RegistrationToken   string
	Email               string
	Password            string
	PasswordConfirmation string
	Nickname            string
	FirstName           string // optional override; empty → from token
	LastName            string // optional override; empty → from token
}

// CompleteTelegramRegistration creates exactly one players row after OIDC registration.
func (uc *AuthUseCase) CompleteTelegramRegistration(ctx context.Context, req CompleteTelegramRegistrationRequest) (*domain.AuthTokens, error) {
	if uc.challenge == nil {
		return nil, ErrInvalidRegistrationToken
	}

	identity, err := uc.challenge.VerifyRegistrationToken(req.RegistrationToken)
	if err != nil {
		return nil, err
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password, req.PasswordConfirmation); err != nil {
		return nil, err
	}
	nickname := strings.TrimSpace(req.Nickname)
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}

	// Ensure Telegram ID is still free.
	if _, err := uc.repos.Players.GetByTgUserID(ctx, identity.UserID); err == nil {
		return nil, ErrTelegramAlreadyRegistered
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to lookup player by tg_user_id: %w", err)
	}

	if _, err := uc.repos.Players.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to lookup player by email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	hashStr := string(hash)
	nick := nickname
	tgUserName := optionalNonEmpty(identity.Username)

	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" {
		firstName = identity.FirstName
	}
	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		lastName = identity.LastName
	}

	player := &domain.Player{
		FirstName:  firstName,
		LastName:   lastName,
		Nickname:   &nick,
		TgUserName: tgUserName,
		Email:      &email,
		Password:   &hashStr,
		TgUserID:   &identity.UserID,
	}

	playerID, err := uc.repos.Players.Create(ctx, player)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			// Race on tg_user_id or email.
			if existing, getErr := uc.repos.Players.GetByTgUserID(ctx, identity.UserID); getErr == nil && existing != nil {
				return nil, ErrTelegramAlreadyRegistered
			}
			if existing, getErr := uc.repos.Players.GetByEmail(ctx, email); getErr == nil && existing != nil {
				return nil, ErrEmailAlreadyExists
			}
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create player: %w", err)
	}
	player.ID = playerID

	uc.log.Info("player registered via telegram web",
		"player_id", player.ID,
		"tg_user_id", identity.UserID,
	)

	return uc.issueTokens(ctx, &domain.AuthenticatedUser{
		PlayerID: player.ID,
		TgUserID: player.TgUserID,
		Role:     "member",
	})
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password, confirmation string) error {
	if utf8.RuneCountInString(password) < minPasswordLength {
		return ErrInvalidPassword
	}
	if password != confirmation {
		return ErrInvalidPassword
	}
	return nil
}

func validateNickname(nickname string) error {
	n := utf8.RuneCountInString(nickname)
	if n < minNicknameLength || n > maxNicknameLength {
		return ErrInvalidNickname
	}
	return nil
}

func (uc *AuthUseCase) syncTgUserName(ctx context.Context, player *domain.Player, tgUserName *string) error {
	current := player.TgUserNameOrEmpty()
	next := ""
	if tgUserName != nil {
		next = *tgUserName
	}
	if current == next {
		player.TgUserName = tgUserName
		return nil
	}
	if err := uc.repos.Players.UpdateTgUserName(ctx, player.ID, tgUserName); err != nil {
		return err
	}
	player.TgUserName = tgUserName
	return nil
}

func optionalNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// RefreshAccessToken issues a new access token using a valid refresh token.
func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := uc.jwt.HashToken(refreshToken)
	stored, err := uc.repos.RefreshTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if stored.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

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
		return nil
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

// UpdateProfileRequest holds editable profile fields.
type UpdateProfileRequest struct {
	FirstName   string
	LastName    string
	Nickname    string
	Email       *string // nil = leave unchanged; empty string clears? We treat empty as invalid if provided
	PhoneNumber *string
}

// UpdateProfile updates personal data for the authenticated player.
// Does not allow changing tg_user_id. Nickname is editable Poker Club nickname.
func (uc *AuthUseCase) UpdateProfile(ctx context.Context, playerID int64, req UpdateProfileRequest) (*domain.Player, error) {
	player, err := uc.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)
	if firstName == "" {
		firstName = player.FirstName
	}
	if lastName == "" {
		lastName = player.LastName
	}

	nickname := strings.TrimSpace(req.Nickname)
	var nicknamePtr *string
	if nickname != "" {
		if err := validateNickname(nickname); err != nil {
			return nil, err
		}
		nicknamePtr = &nickname
	}

	email := player.Email
	if req.Email != nil {
		trimmed := strings.TrimSpace(strings.ToLower(*req.Email))
		if trimmed == "" {
			email = nil
		} else {
			if err := validateEmail(trimmed); err != nil {
				return nil, err
			}
			if player.Email == nil || *player.Email != trimmed {
				if existing, err := uc.repos.Players.GetByEmail(ctx, trimmed); err == nil && existing.ID != playerID {
					return nil, ErrEmailAlreadyExists
				} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
					return nil, fmt.Errorf("failed to check email: %w", err)
				}
			}
			email = &trimmed
		}
	}

	phone := player.PhoneNumber
	if req.PhoneNumber != nil {
		trimmed := strings.TrimSpace(*req.PhoneNumber)
		if trimmed == "" {
			phone = nil
		} else {
			phone = &trimmed
		}
	}

	if err := uc.repos.Players.UpdateProfile(ctx, playerID, firstName, lastName, nicknamePtr, email, phone); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return uc.repos.Players.GetByID(ctx, playerID)
}

// ChangePasswordRequest sets or changes the player's password.
type ChangePasswordRequest struct {
	CurrentPassword      string // required when player already has a password
	NewPassword          string
	PasswordConfirmation string
}

// ChangePassword sets or updates the password for the authenticated player.
func (uc *AuthUseCase) ChangePassword(ctx context.Context, playerID int64, req ChangePasswordRequest) error {
	player, err := uc.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return err
	}

	if err := validatePassword(req.NewPassword, req.PasswordConfirmation); err != nil {
		return err
	}

	if player.HasPassword() {
		if req.CurrentPassword == "" {
			return ErrCurrentPasswordRequired
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*player.Password), []byte(req.CurrentPassword)); err != nil {
			return ErrCurrentPasswordInvalid
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return uc.repos.Players.UpdatePassword(ctx, playerID, string(hash))
}

// AccessTokenTTL returns the configured access token lifetime.
func (uc *AuthUseCase) AccessTokenTTL() time.Duration {
	return uc.jwt.AccessTokenTTL()
}

func (uc *AuthUseCase) issueTokens(ctx context.Context, user *domain.AuthenticatedUser) (*domain.AuthTokens, error) {
	accessToken, err := uc.jwt.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

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
	ErrInvalidCredentials      = &AuthError{Code: "INVALID_CREDENTIALS", Msg: "invalid credentials"}
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

// HTTPStatusHint returns a suggested HTTP status for the auth error.
func (e *AuthError) HTTPStatusHint() int {
	switch e.Code {
	case "EMAIL_ALREADY_EXISTS", "TELEGRAM_ALREADY_REGISTERED":
		return 409
	case "INVALID_EMAIL", "INVALID_PASSWORD", "INVALID_NICKNAME", "VALIDATION_ERROR",
		"CURRENT_PASSWORD_REQUIRED", "CURRENT_PASSWORD_INVALID":
		return 400
	case "TELEGRAM_REGISTRATION_REQUIRED":
		return 200
	default:
		return 401
	}
}

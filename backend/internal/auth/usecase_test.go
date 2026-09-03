package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"poker-club/backend/internal/domain"
)

// testLogger creates a logger for tests.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- Mock repositories ---

type mockPlayerRepo struct {
	player     *domain.Player
	getErr     error
	createID   int64
	createErr  error
	created    *domain.Player
	lastSeenID int64
}

func (m *mockPlayerRepo) Ping(ctx context.Context) error { return nil }

func (m *mockPlayerRepo) Create(ctx context.Context, player *domain.Player) (int64, error) {
	m.created = player
	return m.createID, m.createErr
}

func (m *mockPlayerRepo) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	if m.player != nil && m.player.ID == id {
		return m.player, nil
	}
	return nil, errors.New("player not found")
}

func (m *mockPlayerRepo) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	if m.player != nil && m.player.TgUserID != nil && *m.player.TgUserID == tgUserID {
		return m.player, nil
	}
	return nil, m.getErr
}

func (m *mockPlayerRepo) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	if m.player != nil && m.player.Nickname == nickname {
		return m.player, nil
	}
	return nil, m.getErr
}

func (m *mockPlayerRepo) UpdateLastSeen(ctx context.Context, id int64) error {
	m.lastSeenID = id
	return nil
}

type mockRefreshTokenRepo struct {
	tokens      map[string]*domain.RefreshToken
	createErr   error
	getErr      error
	revokeErr   error
	revokeAllID int64
}

func (m *mockRefreshTokenRepo) Ping(ctx context.Context) error { return nil }

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) (int64, error) {
	if m.tokens == nil {
		m.tokens = make(map[string]*domain.RefreshToken)
	}
	m.tokens[token.TokenHash] = token
	return 1, m.createErr
}

func (m *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	t, ok := m.tokens[tokenHash]
	if !ok {
		return nil, errors.New("refresh token not found")
	}
	return t, nil
}

func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	if t, ok := m.tokens[tokenHash]; ok {
		now := time.Now()
		t.RevokedAt = &now
	}
	return m.revokeErr
}

func (m *mockRefreshTokenRepo) RevokeByPlayer(ctx context.Context, playerID int64) error {
	m.revokeAllID = playerID
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockClubMemberRepo struct{}

func (m *mockClubMemberRepo) Ping(ctx context.Context) error { return nil }
func (m *mockClubMemberRepo) Create(ctx context.Context, member *domain.ClubMember) (int64, error) {
	return 0, nil
}
func (m *mockClubMemberRepo) GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*domain.ClubMember, error) {
	return nil, errors.New("not found")
}
func (m *mockClubMemberRepo) GetByClubWithPlayers(ctx context.Context, clubID int64) ([]*domain.ClubMemberWithPlayer, error) {
	return nil, nil
}
func (m *mockClubMemberRepo) CountActiveMembers(ctx context.Context, clubID int64) (int, error) {
	return 0, nil
}
func (m *mockClubMemberRepo) UpdateRole(ctx context.Context, clubID, playerID int64, role string) error {
	return nil
}
func (m *mockClubMemberRepo) UpdateStatus(ctx context.Context, clubID, playerID int64, status string) error {
	return nil
}
func (m *mockClubMemberRepo) UpdateAccepted(ctx context.Context, clubID, playerID int64, accepted bool) error {
	return nil
}

// --- Tests ---

func TestLogin_InvalidCredentials_PlayerNotFound(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			getErr: errors.New("player not found"),
		},
		ClubMembers: &mockClubMemberRepo{},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	_, err := uc.Login(context.Background(), "nonexistent", "password")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{
				ID:       1,
				Nickname: "testuser",
				Password: string(hashedPassword),
			},
		},
		ClubMembers: &mockClubMemberRepo{},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	_, err := uc.Login(context.Background(), "testuser", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{
				ID:       1,
				Nickname: "testuser",
				Password: string(hashedPassword),
			},
		},
		ClubMembers: &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	tokens, err := uc.Login(context.Background(), "testuser", "correctpassword")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestJWTManager_GenerateAndValidateAccessToken(t *testing.T) {
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)

	user := &domain.AuthenticatedUser{
		PlayerID: 1,
		TgUserID: ptrInt64(12345),
		Role:     "member",
	}

	token, err := jwt.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	claims, err := jwt.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.PlayerID != 1 {
		t.Errorf("expected player ID 1, got %d", claims.PlayerID)
	}
	if claims.TgUserID == nil || *claims.TgUserID != 12345 {
		t.Errorf("expected tg_user_id 12345, got %v", claims.TgUserID)
	}
	if claims.Role != "member" {
		t.Errorf("expected role 'member', got %q", claims.Role)
	}
}

func TestJWTManager_ValidateAccessToken_Expired(t *testing.T) {
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", -time.Hour, 24*time.Hour)

	user := &domain.AuthenticatedUser{
		PlayerID: 1,
		Role:     "member",
	}

	token, err := jwt.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	_, err = jwt.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestJWTManager_ValidateAccessToken_InvalidSignature(t *testing.T) {
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)

	user := &domain.AuthenticatedUser{
		PlayerID: 1,
		Role:     "member",
	}

	token, err := jwt.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// Tamper with the token by replacing the signature part.
	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts in JWT, got %d", len(parts))
	}
	// Replace the signature with an invalid one.
	tampered := parts[0] + "." + parts[1] + ".invalidsignature"

	_, err = jwt.ValidateAccessToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)

	token1, err := jwt.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if token1 == "" {
		t.Fatal("expected non-empty refresh token")
	}

	token2, err := jwt.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected different refresh tokens")
	}
}

func TestJWTManager_HashToken(t *testing.T) {
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)

	token := "test-refresh-token"
	hash1 := jwt.HashToken(token)
	hash2 := jwt.HashToken(token)

	if hash1 != hash2 {
		t.Error("hash should be deterministic for the same token")
	}

	if hash1 == token {
		t.Error("hash should not equal the raw token")
	}
}

func TestValidateTelegramInitData_EmptyData(t *testing.T) {
	_, err := ValidateTelegramInitData("", "test-bot-token", time.Hour)
	if err != ErrInvalidInitData {
		t.Errorf("expected ErrInvalidInitData, got %v", err)
	}
}

func TestValidateTelegramInitData_InvalidData(t *testing.T) {
	_, err := ValidateTelegramInitData("invalid-data", "test-bot-token", time.Hour)
	if err != ErrInvalidInitData {
		t.Errorf("expected ErrInvalidInitData, got %v", err)
	}
}

func TestValidateTelegramInitData_InvalidSignature(t *testing.T) {
	// Build a fake initData with invalid signature.
	initData := "user=%7B%22id%22%3A12345%2C%22first_name%22%3A%22Test%22%7D&auth_date=1234567890&hash=invalidhash"
	_, err := ValidateTelegramInitData(initData, "test-bot-token", 24*time.Hour)
	if err != ErrInvalidInitData {
		t.Errorf("expected ErrInvalidInitData, got %v", err)
	}
}

func TestValidateTelegramInitData_Expired(t *testing.T) {
	// Build a valid initData with an old auth_date.
	// We need to compute the correct hash for this test.
	botToken := "test-bot-token"
	authDate := time.Now().Add(-25 * time.Hour).Unix()

	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser"}`
	userEncoded := url.QueryEscape(userJSON)

	dataCheckString := fmt.Sprintf("auth_date=%d\nuser=%s", authDate, userJSON)

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", userEncoded, authDate, hash)

	_, err := ValidateTelegramInitData(initData, botToken, time.Hour)
	if err != ErrExpiredInitData {
		t.Errorf("expected ErrExpiredInitData, got %v", err)
	}
}

func TestValidateTelegramInitData_Valid(t *testing.T) {
	botToken := "test-bot-token"
	authDate := time.Now().Unix()

	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser"}`
	userEncoded := url.QueryEscape(userJSON)

	dataCheckString := fmt.Sprintf("auth_date=%d\nuser=%s", authDate, userJSON)

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", userEncoded, authDate, hash)

	result, err := ValidateTelegramInitData(initData, botToken, 24*time.Hour)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.UserID != 12345 {
		t.Errorf("expected user ID 12345, got %d", result.UserID)
	}
	if result.FirstName != "Test" {
		t.Errorf("expected first name 'Test', got %q", result.FirstName)
	}
	if result.LastName != "User" {
		t.Errorf("expected last name 'User', got %q", result.LastName)
	}
	if result.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", result.Username)
	}
}

func TestAuthenticateTelegram_NewPlayer(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			getErr:  errors.New("player not found"),
			createID: 1,
		},
		ClubMembers:    &mockClubMemberRepo{},
		RefreshTokens:  &mockRefreshTokenRepo{},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	// Build valid initData.
	botToken := "test-bot-token"
	authDate := time.Now().Unix()
	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser"}`
	userEncoded := url.QueryEscape(userJSON)
	dataCheckString := fmt.Sprintf("auth_date=%d\nuser=%s", authDate, userJSON)
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))
	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", userEncoded, authDate, hash)

	tokens, err := uc.AuthenticateTelegram(context.Background(), initData)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	// Verify player was created.
	playerRepo := repos.Players.(*mockPlayerRepo)
	if playerRepo.created == nil {
		t.Fatal("expected player to be created")
	}
	if playerRepo.created.FirstName != "Test" {
		t.Errorf("expected first name 'Test', got %q", playerRepo.created.FirstName)
	}
	if playerRepo.created.TgUserID == nil || *playerRepo.created.TgUserID != 12345 {
		t.Errorf("expected tg_user_id 12345, got %v", playerRepo.created.TgUserID)
	}
}

func TestAuthenticateTelegram_ExistingPlayer(t *testing.T) {
	existingPlayer := &domain.Player{
		ID:       1,
		Nickname: "testuser",
		TgUserID: ptrInt64(12345),
	}
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: existingPlayer,
		},
		ClubMembers:   &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	// Build valid initData.
	botToken := "test-bot-token"
	authDate := time.Now().Unix()
	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser"}`
	userEncoded := url.QueryEscape(userJSON)
	dataCheckString := fmt.Sprintf("auth_date=%d\nuser=%s", authDate, userJSON)
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))
	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", userEncoded, authDate, hash)

	tokens, err := uc.AuthenticateTelegram(context.Background(), initData)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{
				ID:       1,
				Nickname: "testuser",
				TgUserID: ptrInt64(12345),
			},
		},
		ClubMembers: &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{
			tokens: make(map[string]*domain.RefreshToken),
		},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	// Generate a refresh token and store its hash.
	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	tokenHash := jwt.HashToken(refreshToken)

	rt := &domain.RefreshToken{
		PlayerID:  1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	repos.RefreshTokens.(*mockRefreshTokenRepo).tokens[tokenHash] = rt

	tokens, err := uc.RefreshAccessToken(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{ID: 1, Nickname: "testuser"},
		},
		ClubMembers:   &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{tokens: make(map[string]*domain.RefreshToken)},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	_, err := uc.RefreshAccessToken(context.Background(), "invalid-token")
	if err != ErrInvalidRefreshToken {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshAccessToken_ExpiredToken(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{ID: 1, Nickname: "testuser"},
		},
		ClubMembers: &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{
			tokens: make(map[string]*domain.RefreshToken),
		},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	refreshToken, _ := jwt.GenerateRefreshToken()
	tokenHash := jwt.HashToken(refreshToken)

	rt := &domain.RefreshToken{
		PlayerID:  1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-time.Hour), // expired
	}
	repos.RefreshTokens.(*mockRefreshTokenRepo).tokens[tokenHash] = rt

	_, err := uc.RefreshAccessToken(context.Background(), refreshToken)
	if err != ErrRefreshTokenExpired {
		t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: &domain.Player{ID: 1, Nickname: "testuser"},
		},
		ClubMembers: &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{
			tokens: make(map[string]*domain.RefreshToken),
		},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	refreshToken, _ := jwt.GenerateRefreshToken()
	tokenHash := jwt.HashToken(refreshToken)

	rt := &domain.RefreshToken{
		PlayerID:  1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	repos.RefreshTokens.(*mockRefreshTokenRepo).tokens[tokenHash] = rt

	err := uc.Logout(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	// Verify token was revoked.
	stored := repos.RefreshTokens.(*mockRefreshTokenRepo).tokens[tokenHash]
	if stored.RevokedAt == nil {
		t.Error("expected token to be revoked")
	}
}

func TestGetCurrentUser_Success(t *testing.T) {
	player := &domain.Player{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "testuser",
		TgUserID:  ptrInt64(12345),
	}
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			player: player,
		},
	}
	jwt := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwt, "test-bot-token", testLogger())

	result, err := uc.GetCurrentUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected player ID 1, got %d", result.ID)
	}
	if result.FirstName != "Test" {
		t.Errorf("expected first name 'Test', got %q", result.FirstName)
	}
}

// Helper function
func ptrInt64(v int64) *int64 {
	return &v
}

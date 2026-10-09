package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"poker-club/backend/internal/domain"
)

func TestAuthenticateTelegram_DBErrorDoesNotCreatePlayer(t *testing.T) {
	dbErr := errors.New("connection refused")
	repos := &domain.Repositories{
		Players: &mockPlayerRepo{
			getErr:   dbErr,
			createID: 99,
		},
		ClubMembers:   &mockClubMemberRepo{},
		RefreshTokens: &mockRefreshTokenRepo{},
	}
	jwtMgr := NewJWTManager("test-secret", "test-issuer", "test-audience", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwtMgr, "test-bot-token", testLogger(), nil, nil)

	botToken := "test-bot-token"
	authDate := time.Now().Unix()
	userJSON := `{"id":12345,"first_name":"Test","last_name":"User","username":"testuser"}`
	userEncoded := url.QueryEscape(userJSON)
	dataCheckString := fmt.Sprintf("auth_date=%d\nuser=%s", authDate, userJSON)
	hash := generateValidTelegramHash(dataCheckString, botToken)
	initData := fmt.Sprintf("user=%s&auth_date=%d&hash=%s", userEncoded, authDate, hash)

	_, err := uc.AuthenticateTelegram(context.Background(), initData)
	if err == nil {
		t.Fatal("expected error on DB failure, got nil")
	}

	playerRepo := repos.Players.(*mockPlayerRepo)
	if playerRepo.created != nil {
		t.Fatal("player must not be created when GetByTgUserID returns a DB error")
	}
}

func TestChallenge_CreateAndVerify(t *testing.T) {
	m := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	ch, err := m.CreateChallenge()
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	if ch.Nonce == "" || ch.ChallengeToken == "" {
		t.Fatal("expected nonce and challenge token")
	}
	nonce, err := m.VerifyChallenge(ch.ChallengeToken)
	if err != nil {
		t.Fatalf("VerifyChallenge: %v", err)
	}
	if nonce != ch.Nonce {
		t.Fatalf("nonce mismatch: got %q want %q", nonce, ch.Nonce)
	}
}

func TestChallenge_TamperedToken(t *testing.T) {
	m := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	ch, err := m.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	tampered := ch.ChallengeToken + "x"
	if _, err := m.VerifyChallenge(tampered); err != ErrTelegramOIDCInvalidNonce {
		t.Fatalf("expected ErrTelegramOIDCInvalidNonce, got %v", err)
	}
}

func TestRegistrationToken_RoundTrip(t *testing.T) {
	m := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	identity := &TelegramOIDCIdentity{
		UserID:    42,
		FirstName: "Ada",
		LastName:  "Lovelace",
		Username:  "ada",
	}
	token, err := m.CreateRegistrationToken(identity)
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.VerifyRegistrationToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != 42 || got.Username != "ada" {
		t.Fatalf("unexpected identity: %+v", got)
	}
}

func TestRegistrationToken_Expired(t *testing.T) {
	m := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	now := time.Now()
	claims := registrationClaims{
		Type:      registrationTokenType,
		TgUserID:  1,
		FirstName: "A",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.VerifyRegistrationToken(signed); err != ErrRegistrationTokenExpired {
		t.Fatalf("expected ErrRegistrationTokenExpired, got %v", err)
	}
}

func TestAuthenticateTelegramWeb_ExistingPlayer(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "test-client"
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	ch, err := challengeMgr.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}

	idToken := signTestIDToken(t, priv, clientID, "999", ch.Nonce, "Ann", "Lee", "annlee", time.Now().Add(time.Hour))

	existing := &domain.Player{
		ID:         7,
		TgUserID:   ptrInt64(999),
		TgUserName: ptrString("oldname"),
	}
	repos := &domain.Repositories{
		Players:       &mockPlayerRepo{player: existing},
		RefreshTokens: &mockRefreshTokenRepo{},
	}
	jwtMgr := NewJWTManager("test-secret", "iss", "aud", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwtMgr, "bot", testLogger(), oidc, challengeMgr)

	result, err := uc.AuthenticateTelegramWeb(context.Background(), idToken, ch.ChallengeToken)
	if err != nil {
		t.Fatalf("AuthenticateTelegramWeb: %v", err)
	}
	if result.RegistrationRequired {
		t.Fatal("expected login, got registration_required")
	}
	if result.Tokens == nil || result.Tokens.AccessToken == "" {
		t.Fatal("expected tokens")
	}
	if existing.TgUserNameOrEmpty() != "annlee" {
		t.Fatalf("expected tg_user_name synced to annlee, got %q", existing.TgUserNameOrEmpty())
	}
}

func TestAuthenticateTelegramWeb_NewPlayerRequiresRegistration(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "test-client"
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	ch, err := challengeMgr.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	idToken := signTestIDToken(t, priv, clientID, "1001", ch.Nonce, "New", "User", "newbie", time.Now().Add(time.Hour))

	repos := &domain.Repositories{
		Players: &mockPlayerRepo{getErr: domain.ErrNotFound},
	}
	jwtMgr := NewJWTManager("test-secret", "iss", "aud", time.Hour, 24*time.Hour)
	uc := NewAuthUseCase(repos, jwtMgr, "bot", testLogger(), oidc, challengeMgr)

	result, err := uc.AuthenticateTelegramWeb(context.Background(), idToken, ch.ChallengeToken)
	if err != nil {
		t.Fatalf("AuthenticateTelegramWeb: %v", err)
	}
	if !result.RegistrationRequired || result.RegistrationToken == "" {
		t.Fatal("expected registration_required with token")
	}
	if result.Tokens != nil {
		t.Fatal("must not issue tokens for new user")
	}
	if repos.Players.(*mockPlayerRepo).created != nil {
		t.Fatal("must not create player before registration completes")
	}
}

func TestAuthenticateTelegramWeb_InvalidNonce(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "test-client"
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	ch, err := challengeMgr.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	idToken := signTestIDToken(t, priv, clientID, "1", "wrong-nonce", "A", "B", "", time.Now().Add(time.Hour))

	uc := NewAuthUseCase(&domain.Repositories{Players: &mockPlayerRepo{}}, NewJWTManager("s", "i", "a", time.Hour, time.Hour), "bot", testLogger(), oidc, challengeMgr)
	_, err = uc.AuthenticateTelegramWeb(context.Background(), idToken, ch.ChallengeToken)
	if err != ErrTelegramOIDCInvalidNonce {
		t.Fatalf("expected ErrTelegramOIDCInvalidNonce, got %v", err)
	}
}

func TestAuthenticateTelegramWeb_ExpiredToken(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "test-client"
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	ch, err := challengeMgr.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	idToken := signTestIDToken(t, priv, clientID, "1", ch.Nonce, "A", "B", "", time.Now().Add(-time.Hour))

	uc := NewAuthUseCase(&domain.Repositories{Players: &mockPlayerRepo{}}, NewJWTManager("s", "i", "a", time.Hour, time.Hour), "bot", testLogger(), oidc, challengeMgr)
	_, err = uc.AuthenticateTelegramWeb(context.Background(), idToken, ch.ChallengeToken)
	if err != ErrTelegramOIDCExpired {
		t.Fatalf("expected ErrTelegramOIDCExpired, got %v", err)
	}
}

func TestAuthenticateTelegramWeb_WrongAudience(t *testing.T) {
	priv := mustRSAKey(t)
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	oidc := NewTelegramOIDCValidator("expected-client", defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	ch, err := challengeMgr.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	idToken := signTestIDToken(t, priv, "other-client", "1", ch.Nonce, "A", "B", "", time.Now().Add(time.Hour))

	uc := NewAuthUseCase(&domain.Repositories{Players: &mockPlayerRepo{}}, NewJWTManager("s", "i", "a", time.Hour, time.Hour), "bot", testLogger(), oidc, challengeMgr)
	_, err = uc.AuthenticateTelegramWeb(context.Background(), idToken, ch.ChallengeToken)
	if !errors.Is(err, ErrTelegramOIDCInvalid) {
		t.Fatalf("expected ErrTelegramOIDCInvalid, got %v", err)
	}
}

func TestCompleteTelegramRegistration_Success(t *testing.T) {
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	regToken, err := challengeMgr.CreateRegistrationToken(&TelegramOIDCIdentity{
		UserID: 55, FirstName: "Reg", LastName: "User", Username: "reguser",
	})
	if err != nil {
		t.Fatal(err)
	}

	repo := &mockPlayerRepo{getErr: domain.ErrNotFound, createID: 10}
	repos := &domain.Repositories{
		Players:       repo,
		RefreshTokens: &mockRefreshTokenRepo{},
	}
	uc := NewAuthUseCase(repos, NewJWTManager("s", "i", "a", time.Hour, 24*time.Hour), "bot", testLogger(), nil, challengeMgr)

	tokens, err := uc.CompleteTelegramRegistration(context.Background(), CompleteTelegramRegistrationRequest{
		RegistrationToken:    regToken,
		Email:                "reg@example.com",
		Password:             "password12345",
		PasswordConfirmation: "password12345",
		Nickname:             "ClubNick",
	})
	if err != nil {
		t.Fatalf("CompleteTelegramRegistration: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if repo.created == nil {
		t.Fatal("expected player create")
	}
	if repo.created.Nickname == nil || *repo.created.Nickname != "ClubNick" {
		t.Fatalf("nickname = %v", repo.created.Nickname)
	}
	if repo.created.TgUserName == nil || *repo.created.TgUserName != "reguser" {
		t.Fatalf("tg_user_name = %v", repo.created.TgUserName)
	}
	if repo.created.Password == nil {
		t.Fatal("expected password hash")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*repo.created.Password), []byte("password12345")); err != nil {
		t.Fatal("password not bcrypt-hashed correctly")
	}
}

func TestCompleteTelegramRegistration_DuplicateEmail(t *testing.T) {
	challengeMgr := NewChallengeManager("test-secret", time.Minute, 15*time.Minute)
	regToken, _ := challengeMgr.CreateRegistrationToken(&TelegramOIDCIdentity{UserID: 55, FirstName: "R"})

	email := "taken@example.com"
	repo := &mockPlayerRepo{
		player: &domain.Player{ID: 1, Email: &email},
		// GetByTgUserID → not found (getErr only used when player doesn't match)
	}
	// Override: first GetByTgUserID should be not found; GetByEmail finds existing.
	// Our mock returns player when email matches for GetByEmail; GetByTgUserID returns not found if tg id doesn't match.
	repo.player.TgUserID = ptrInt64(1) // different from 55

	uc := NewAuthUseCase(&domain.Repositories{Players: repo}, NewJWTManager("s", "i", "a", time.Hour, time.Hour), "bot", testLogger(), nil, challengeMgr)
	_, err := uc.CompleteTelegramRegistration(context.Background(), CompleteTelegramRegistrationRequest{
		RegistrationToken:    regToken,
		Email:                email,
		Password:             "password12345",
		PasswordConfirmation: "password12345",
		Nickname:             "Nick",
	})
	if err != ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestValidateIDToken_InvalidSignature(t *testing.T) {
	priv := mustRSAKey(t)
	other := mustRSAKey(t)
	clientID := "c"
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &other.PublicKey, nil
	}
	token := signTestIDToken(t, priv, clientID, "1", "n", "A", "B", "", time.Now().Add(time.Hour))
	_, err := oidc.ValidateIDToken(context.Background(), token, "n")
	if !errors.Is(err, ErrTelegramOIDCInvalid) {
		t.Fatalf("expected ErrTelegramOIDCInvalid, got %v", err)
	}
}

func TestValidateIDToken_UsesTelegramIDClaimNotSub(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "client-1"
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	tgUserID := int64(987654321)
	nonce := "nonce-abc"
	token := signTestIDToken(t, priv, clientID, "987654321", nonce, "John", "Doe", "johndoe", time.Now().Add(time.Hour))

	identity, err := oidc.ValidateIDToken(context.Background(), token, nonce)
	if err != nil {
		t.Fatalf("ValidateIDToken: %v", err)
	}
	if identity.UserID != tgUserID {
		t.Fatalf("UserID = %d, want Telegram claim id %d (must not use OIDC sub)", identity.UserID, tgUserID)
	}
	if identity.FirstName != "John" || identity.Username != "johndoe" {
		t.Fatalf("unexpected profile claims: %+v", identity)
	}
}

func TestValidateIDToken_AcceptsStringIDAndNumericAud(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "8661264398"
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	// Build token payload manually so id/aud are JSON strings/numbers like Telegram.
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payloadObj := map[string]any{
		"iss":                defaultTelegramIssuer,
		"aud":                8661264398, // numeric aud
		"sub":                "oidc-sub-different",
		"id":                 "987654321", // string id
		"nonce":              "n1",
		"given_name":         "Ada",
		"family_name":        "Lovelace",
		"preferred_username": "ada",
		"exp":                time.Now().Add(time.Hour).Unix(),
		"iat":                time.Now().Unix(),
	}
	payloadJSON, err := json.Marshal(payloadObj)
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)

	identity, err := oidc.ValidateIDToken(context.Background(), token, "n1")
	if err != nil {
		t.Fatalf("ValidateIDToken: %v", err)
	}
	if identity.UserID != 987654321 {
		t.Fatalf("UserID = %d, want 987654321", identity.UserID)
	}
}

func TestValidateIDToken_MissingTelegramID(t *testing.T) {
	priv := mustRSAKey(t)
	clientID := "client-1"
	oidc := NewTelegramOIDCValidator(clientID, defaultTelegramIssuer, "")
	oidc.keyFuncOverride = func(token *jwt.Token) (interface{}, error) {
		return &priv.PublicKey, nil
	}

	claims := telegramOIDCClaims{
		Nonce: "n",
		// TelegramID intentionally omitted (0)
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    defaultTelegramIssuer,
			Subject:   "1234123412341234123",
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := tok.SignedString(priv)
	if err != nil {
		t.Fatal(err)
	}

	_, err = oidc.ValidateIDToken(context.Background(), signed, "n")
	if !errors.Is(err, ErrTelegramOIDCInvalid) {
		t.Fatalf("expected ErrTelegramOIDCInvalid when id claim missing, got %v", err)
	}
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// signTestIDToken builds a Telegram-like id_token.
// telegramUserID is claim "id" (maps to players.tg_user_id).
// OIDC "sub" is intentionally a different value to catch regressions that
// mistakenly use sub for Telegram identity.
func signTestIDToken(t *testing.T, priv *rsa.PrivateKey, aud, tgUserIDStr, nonce, given, family, username string, exp time.Time) string {
	t.Helper()
	tgID, err := strconv.ParseInt(tgUserIDStr, 10, 64)
	if err != nil {
		t.Fatalf("tgUserIDStr: %v", err)
	}
	claims := telegramOIDCClaims{
		Nonce:             nonce,
		TelegramID:        telegramUserID(tgID),
		GivenName:         given,
		FamilyName:        family,
		PreferredUsername: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    defaultTelegramIssuer,
			Subject:   "oidc-sub-not-tg-id-" + tgUserIDStr,
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

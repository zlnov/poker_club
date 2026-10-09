package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultChallengeTTL    = 5 * time.Minute
	defaultRegistrationTTL = 15 * time.Minute
	challengeTokenType     = "tg_web_challenge"
	registrationTokenType  = "tg_web_registration"
)

// ChallengeManager issues and verifies signed challenge / registration tokens.
type ChallengeManager struct {
	secret           []byte
	challengeTTL     time.Duration
	registrationTTL  time.Duration
}

// NewChallengeManager creates a ChallengeManager using the app JWT secret.
func NewChallengeManager(secret string, challengeTTL, registrationTTL time.Duration) *ChallengeManager {
	if challengeTTL <= 0 {
		challengeTTL = defaultChallengeTTL
	}
	if registrationTTL <= 0 {
		registrationTTL = defaultRegistrationTTL
	}
	return &ChallengeManager{
		secret:          []byte(secret),
		challengeTTL:    challengeTTL,
		registrationTTL: registrationTTL,
	}
}

type challengeClaims struct {
	Type  string `json:"typ"`
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}

type registrationClaims struct {
	Type       string  `json:"typ"`
	TgUserID   int64   `json:"tg_user_id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	TgUserName *string `json:"tg_user_name,omitempty"`
	jwt.RegisteredClaims
}

// ChallengeResult is returned to the frontend before Telegram Login popup.
type ChallengeResult struct {
	ChallengeToken string
	Nonce          string
	ExpiresIn      int
}

// CreateChallenge generates a cryptographically random nonce bound in a signed token.
func (m *ChallengeManager) CreateChallenge() (*ChallengeResult, error) {
	nonce, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	jti, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate jti: %w", err)
	}

	now := time.Now()
	claims := challengeClaims{
		Type:  challengeTokenType,
		Nonce: nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.challengeTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	return &ChallengeResult{
		ChallengeToken: signed,
		Nonce:          nonce,
		ExpiresIn:      int(m.challengeTTL.Seconds()),
	}, nil
}

// VerifyChallenge validates the challenge token and returns the bound nonce.
func (m *ChallengeManager) VerifyChallenge(challengeToken string) (nonce string, err error) {
	claims := &challengeClaims{}
	token, err := jwt.ParseWithClaims(challengeToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTelegramOIDCInvalidNonce
		}
		return "", ErrTelegramOIDCInvalidNonce
	}
	if !token.Valid || claims.Type != challengeTokenType || claims.Nonce == "" {
		return "", ErrTelegramOIDCInvalidNonce
	}
	return claims.Nonce, nil
}

// CreateRegistrationToken signs a short-lived token containing verified Telegram identity.
func (m *ChallengeManager) CreateRegistrationToken(identity *TelegramOIDCIdentity) (string, error) {
	jti, err := randomHex(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate jti: %w", err)
	}

	var tgUserName *string
	if identity.Username != "" {
		u := identity.Username
		tgUserName = &u
	}

	now := time.Now()
	claims := registrationClaims{
		Type:       registrationTokenType,
		TgUserID:   identity.UserID,
		FirstName:  identity.FirstName,
		LastName:   identity.LastName,
		TgUserName: tgUserName,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   fmt.Sprintf("%d", identity.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.registrationTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// VerifyRegistrationToken validates and returns Telegram identity from a registration token.
func (m *ChallengeManager) VerifyRegistrationToken(tokenStr string) (*TelegramOIDCIdentity, error) {
	claims := &registrationClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrRegistrationTokenExpired
		}
		return nil, ErrInvalidRegistrationToken
	}
	if !token.Valid || claims.Type != registrationTokenType || claims.TgUserID == 0 {
		return nil, ErrInvalidRegistrationToken
	}

	username := ""
	if claims.TgUserName != nil {
		username = *claims.TgUserName
	}

	return &TelegramOIDCIdentity{
		UserID:    claims.TgUserID,
		FirstName: claims.FirstName,
		LastName:  claims.LastName,
		Username:  username,
	}, nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

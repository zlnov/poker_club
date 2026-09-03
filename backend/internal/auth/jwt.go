package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"poker-club/backend/internal/domain"
)

// JWTManager handles JWT token creation and validation.
type JWTManager struct {
	secretKey     []byte
	issuer        string
	audience      string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	signingMethod jwt.SigningMethod
}

// NewJWTManager creates a new JWTManager with the given configuration.
func NewJWTManager(secretKey, issuer, audience string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		audience:      audience,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		signingMethod: jwt.SigningMethodHS256,
	}
}

// Claims represents the JWT claims for an authenticated user.
type Claims struct {
	PlayerID int64  `json:"sub"`
	TgUserID *int64 `json:"tg_user_id,omitempty"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT access token for the given user.
func (m *JWTManager) GenerateAccessToken(user *domain.AuthenticatedUser) (string, error) {
	now := time.Now()
	claims := Claims{
		PlayerID: user.PlayerID,
		TgUserID: user.TgUserID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.PlayerID),
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(m.signingMethod, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token string.
// The raw token is returned to the caller; only its hash should be stored.
func (m *JWTManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken hashes a raw refresh token for storage.
// The raw token is never stored in the database.
func (m *JWTManager) HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// ValidateAccessToken validates a JWT access token and returns the claims.
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshTokenTTL returns the refresh token lifetime.
func (m *JWTManager) RefreshTokenTTL() time.Duration {
	return m.refreshTTL
}

// AccessTokenTTL returns the access token lifetime.
func (m *JWTManager) AccessTokenTTL() time.Duration {
	return m.accessTTL
}

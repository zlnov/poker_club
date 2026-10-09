package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTelegramIssuer  = "https://oauth.telegram.org"
	defaultTelegramJWKSURL = "https://oauth.telegram.org/.well-known/jwks.json"
	jwksCacheTTL           = 1 * time.Hour
	jwksHTTPTimeout        = 15 * time.Second
)

// newJWKSHTTPClient returns an HTTP client for fetching Telegram JWKS.
// It dials IPv4 only (tcp4): on some VPN setups AAAA resolves but IPv6 hangs,
// causing Go's default dialer to hit Client.Timeout while curl (Happy Eyeballs / IPv4) succeeds.
func newJWKSHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Timeout:   jwksHTTPTimeout,
		Transport: transport,
	}
}

// TelegramOIDCIdentity is the verified Telegram identity from an OIDC id_token.
type TelegramOIDCIdentity struct {
	UserID    int64
	FirstName string
	LastName  string
	Username  string // preferred_username without @; may be empty
}

// TelegramOIDCValidator validates Telegram Login OIDC id_tokens.
type TelegramOIDCValidator struct {
	clientID string
	issuer   string
	jwksURL  string
	http     *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time

	// keyFuncOverride is used in tests to inject keys without HTTP.
	keyFuncOverride jwt.Keyfunc
}

// NewTelegramOIDCValidator creates a validator for Telegram OIDC id_tokens.
func NewTelegramOIDCValidator(clientID, issuer, jwksURL string) *TelegramOIDCValidator {
	if issuer == "" {
		issuer = defaultTelegramIssuer
	}
	if jwksURL == "" {
		jwksURL = defaultTelegramJWKSURL
	}
	return &TelegramOIDCValidator{
		clientID: clientID,
		issuer:   issuer,
		jwksURL:  jwksURL,
		http:     newJWKSHTTPClient(),
		keys:     make(map[string]*rsa.PublicKey),
	}
}

// telegramUserID unmarshals Telegram profile claim "id" from either a JSON
// number or a JSON string (Telegram has been observed to emit both forms).
type telegramUserID int64

func (id *telegramUserID) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*id = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("telegram id string: %w", err)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*id = 0
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("telegram id string parse: %w", err)
		}
		*id = telegramUserID(n)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("telegram id number: %w", err)
	}
	n, err := num.Int64()
	if err != nil {
		f, ferr := num.Float64()
		if ferr != nil {
			return fmt.Errorf("telegram id number parse: %w", err)
		}
		*id = telegramUserID(int64(f))
		return nil
	}
	*id = telegramUserID(n)
	return nil
}

// telegramOIDCClaims are the claims we accept from Telegram id_token.
// Note: Telegram OIDC separates:
//   - sub  — OIDC subject (openid scope); NOT the Telegram user id
//   - id   — Telegram user id (profile scope); maps to players.tg_user_id
// See https://core.telegram.org/bots/telegram-login
type telegramOIDCClaims struct {
	Nonce             string         `json:"nonce"`
	TelegramID        telegramUserID `json:"id"`
	GivenName         string         `json:"given_name"`
	FamilyName        string         `json:"family_name"`
	PreferredUsername string         `json:"preferred_username"`
	jwt.RegisteredClaims
}

// ValidateIDToken verifies the Telegram OIDC id_token and returns the identity.
// It checks signature (JWKS), iss, aud, exp, and nonce.
// Telegram user id is taken from profile claim "id" (not OIDC "sub").
func (v *TelegramOIDCValidator) ValidateIDToken(ctx context.Context, idToken, expectedNonce string) (*TelegramOIDCIdentity, error) {
	if idToken == "" {
		return nil, ErrTelegramOIDCInvalid
	}
	if expectedNonce == "" {
		return nil, ErrTelegramOIDCInvalidNonce
	}

	// MapClaims: Telegram may emit aud/id as JSON numbers or strings.
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithExpirationRequired(),
	)

	keyFunc := v.keyFuncOverride
	if keyFunc == nil {
		keyFunc = func(token *jwt.Token) (interface{}, error) {
			return v.lookupKey(ctx, token)
		}
	}

	claims := jwt.MapClaims{}
	token, err := parser.ParseWithClaims(idToken, claims, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTelegramOIDCExpired
		}
		if v.keyFuncOverride == nil && isLikelyKeyMiss(err) {
			if refreshErr := v.refreshKeys(ctx, true); refreshErr == nil {
				claims = jwt.MapClaims{}
				token, err = parser.ParseWithClaims(idToken, claims, func(t *jwt.Token) (interface{}, error) {
					return v.lookupKey(ctx, t)
				})
			}
		}
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, ErrTelegramOIDCExpired
			}
			return nil, fmt.Errorf("%w: parse: %v", ErrTelegramOIDCInvalid, err)
		}
	}
	if !token.Valid {
		return nil, ErrTelegramOIDCInvalid
	}

	iss, _ := claims["iss"].(string)
	if iss != v.issuer {
		return nil, fmt.Errorf("%w: issuer mismatch", ErrTelegramOIDCInvalid)
	}
	if !audienceIncludes(claims["aud"], v.clientID) {
		return nil, fmt.Errorf("%w: audience mismatch", ErrTelegramOIDCInvalid)
	}

	nonce, _ := claims["nonce"].(string)
	if nonce != expectedNonce {
		return nil, ErrTelegramOIDCInvalidNonce
	}

	// players.tg_user_id ← claim "id" (profile scope), never OIDC "sub".
	userID, ok := claimAsInt64(claims["id"])
	if !ok || userID == 0 {
		return nil, fmt.Errorf("%w: missing telegram profile claim id", ErrTelegramOIDCInvalid)
	}

	username := strings.TrimPrefix(claimAsString(claims["preferred_username"]), "@")

	return &TelegramOIDCIdentity{
		UserID:    userID,
		FirstName: claimAsString(claims["given_name"]),
		LastName:  claimAsString(claims["family_name"]),
		Username:  username,
	}, nil
}

func audienceIncludes(aud any, clientID string) bool {
	if clientID == "" {
		return false
	}
	switch a := aud.(type) {
	case string:
		return a == clientID
	case float64:
		return strconv.FormatInt(int64(a), 10) == clientID
	case json.Number:
		return a.String() == clientID
	case []any:
		for _, item := range a {
			if audienceIncludes(item, clientID) {
				return true
			}
		}
	}
	return false
}

func claimAsString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatInt(int64(s), 10)
	case json.Number:
		return s.String()
	default:
		return ""
	}
}

func claimAsInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			f, ferr := n.Float64()
			if ferr != nil {
				return 0, false
			}
			return int64(f), true
		}
		return i, true
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, false
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

func isLikelyKeyMiss(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "key is of invalid type") ||
		strings.Contains(msg, "verification error") ||
		strings.Contains(msg, "kid")
}

func (v *TelegramOIDCValidator) lookupKey(ctx context.Context, token *jwt.Token) (*rsa.PublicKey, error) {
	if token.Method.Alg() != jwt.SigningMethodRS256.Name {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	kid, _ := token.Header["kid"].(string)

	if err := v.refreshKeys(ctx, false); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	if kid != "" {
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
		return nil, fmt.Errorf("unknown kid: %s", kid)
	}
	// If no kid, use the single key when only one is present.
	if len(v.keys) == 1 {
		for _, key := range v.keys {
			return key, nil
		}
	}
	return nil, errors.New("no matching JWKS key")
}

func (v *TelegramOIDCValidator) refreshKeys(ctx context.Context, force bool) error {
	v.mu.RLock()
	fresh := !force && len(v.keys) > 0 && time.Since(v.fetchedAt) < jwksCacheTTL
	v.mu.RUnlock()
	if fresh {
		return nil
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock.
	if !force && len(v.keys) > 0 && time.Since(v.fetchedAt) < jwksCacheTTL {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("jwks request: %w", err)
	}
	resp, err := v.http.Do(req)
	if err != nil {
		return fmt.Errorf("jwks fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks status: %d", resp.StatusCode)
	}

	var set jwksSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return fmt.Errorf("jwks decode: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" {
			continue
		}
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		kid := k.Kid
		if kid == "" {
			kid = fmt.Sprintf("key-%d", len(keys))
		}
		keys[kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("jwks contained no usable RSA keys")
	}

	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

type jwksSet struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}
	if eInt == 0 {
		return nil, errors.New("invalid exponent")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

// OIDC-related auth errors (stable codes for frontend).
var (
	ErrTelegramOIDCInvalid      = &AuthError{Code: "TELEGRAM_OIDC_INVALID", Msg: "invalid telegram oidc token"}
	ErrTelegramOIDCExpired      = &AuthError{Code: "TELEGRAM_OIDC_EXPIRED", Msg: "telegram oidc token expired"}
	ErrTelegramOIDCInvalidNonce = &AuthError{Code: "TELEGRAM_OIDC_INVALID_NONCE", Msg: "invalid telegram oidc nonce"}
	ErrTelegramRegistrationRequired = &AuthError{Code: "TELEGRAM_REGISTRATION_REQUIRED", Msg: "registration required"}
	ErrEmailAlreadyExists       = &AuthError{Code: "EMAIL_ALREADY_EXISTS", Msg: "email already exists"}
	ErrInvalidEmail             = &AuthError{Code: "INVALID_EMAIL", Msg: "invalid email"}
	ErrInvalidPassword          = &AuthError{Code: "INVALID_PASSWORD", Msg: "invalid password"}
	ErrInvalidNickname          = &AuthError{Code: "INVALID_NICKNAME", Msg: "invalid nickname"}
	ErrInvalidRegistrationToken = &AuthError{Code: "INVALID_REGISTRATION_TOKEN", Msg: "invalid registration token"}
	ErrRegistrationTokenExpired = &AuthError{Code: "REGISTRATION_TOKEN_EXPIRED", Msg: "registration token expired"}
	ErrTelegramAlreadyRegistered = &AuthError{Code: "TELEGRAM_ALREADY_REGISTERED", Msg: "telegram account already registered"}
	ErrCurrentPasswordRequired   = &AuthError{Code: "CURRENT_PASSWORD_REQUIRED", Msg: "current password required"}
	ErrCurrentPasswordInvalid    = &AuthError{Code: "CURRENT_PASSWORD_INVALID", Msg: "current password invalid"}
)

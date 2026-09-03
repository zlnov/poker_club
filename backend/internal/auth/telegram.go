package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramInitData represents the parsed and validated Telegram initData.
type TelegramInitData struct {
	UserID    int64
	FirstName string
	LastName  string
	Username  string
	AuthDate  int64
	Hash      string
}

// TelegramInitDataError represents an error during Telegram initData validation.
type TelegramInitDataError struct {
	Code string
	Msg  string
}

func (e *TelegramInitDataError) Error() string {
	return e.Msg
}

// TelegramInitData errors
var (
	ErrInvalidInitData = &TelegramInitDataError{Code: "INVALID_TELEGRAM_INIT_DATA", Msg: "invalid telegram init data"}
	ErrExpiredInitData = &TelegramInitDataError{Code: "TELEGRAM_INIT_DATA_EXPIRED", Msg: "telegram init data expired"}
)

// ValidateTelegramInitData validates the Telegram initData string.
// It checks the HMAC signature and expiration.
// Returns the parsed initData on success.
func ValidateTelegramInitData(initData, botToken string, maxAge time.Duration) (*TelegramInitData, error) {
	if initData == "" {
		return nil, ErrInvalidInitData
	}

	// Parse the initData as URL-encoded form data.
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	// Extract the hash parameter.
	hash := values.Get("hash")
	if hash == "" {
		return nil, ErrInvalidInitData
	}

	// Remove hash from values for signature verification.
	values.Del("hash")

	// Build the data-check string: sorted key=value pairs joined by newline.
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckStrings []string
	for _, k := range keys {
		dataCheckStrings = append(dataCheckStrings, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(dataCheckStrings, "\n")

	// Compute the secret key: SHA-256 of the bot token.
	secretKey := sha256.Sum256([]byte(botToken))

	// Compute HMAC-SHA-256 of the data-check string using the secret key.
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Compare the computed hash with the provided hash.
	if !hmac.Equal([]byte(expectedHash), []byte(hash)) {
		return nil, ErrInvalidInitData
	}

	// Parse auth_date and check expiration.
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, ErrInvalidInitData
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	// Check expiration.
	if time.Since(time.Unix(authDate, 0)) > maxAge {
		return nil, ErrExpiredInitData
	}

	// Parse the user data from the "user" parameter.
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, ErrInvalidInitData
	}

	user, err := parseTelegramUser(userJSON)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	return &TelegramInitData{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		AuthDate:  authDate,
		Hash:      hash,
	}, nil
}

// telegramUser represents the user object embedded in initData.
type telegramUser struct {
	UserID    int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// parseTelegramUser parses the JSON-encoded user object from initData.
func parseTelegramUser(jsonStr string) (*telegramUser, error) {
	// Simple JSON parsing without external dependencies.
	// The user field is a JSON object like: {"id":123,"first_name":"John","last_name":"Doe","username":"johndoe"}
	user := &telegramUser{}

	// Use encoding/json for parsing.
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	if err := decoder.Decode(user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	if user.UserID == 0 {
		return nil, errors.New("user id is required")
	}

	return user, nil
}

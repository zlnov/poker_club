package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"poker-club/backend/internal/auth"
	"poker-club/backend/internal/domain"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const authenticatedUserKey contextKey = "authenticated_user"

// CORSMiddleware creates a Gin middleware that handles CORS.
// It allows requests from the configured frontend origin and supports
// credentials (JWT Bearer tokens in the Authorization header).
//
// The middleware:
// - Sets Access-Control-Allow-Origin to the specific frontend origin (not "*")
// - Sets Access-Control-Allow-Credentials to "true"
// - Handles OPTIONS preflight requests by returning 204
// - Allows standard methods (GET, POST, PATCH, DELETE, PUT, OPTIONS)
// - Allows standard headers including Authorization and Content-Type
func CORSMiddleware(frontendOrigins []string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(frontendOrigins))

	for _, origin := range frontendOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Only set CORS headers if the origin matches the configured frontend origin
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		// Handle preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, ngrok-skip-browser-warning")
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// GetAuthenticatedUser retrieves the authenticated user from the Gin context.
// Returns nil and false if no authenticated user is present.
func GetAuthenticatedUser(c *gin.Context) (*domain.AuthenticatedUser, bool) {
	val, exists := c.Get(string(authenticatedUserKey))
	if !exists {
		return nil, false
	}
	user, ok := val.(*domain.AuthenticatedUser)
	return user, ok
}

// AuthMiddleware creates a Gin middleware that validates JWT tokens
// and sets the authenticated user in the request context.
// This middleware handles authentication only, not authorization.
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the token from the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("AUTHENTICATION_REQUIRED", "authentication required"))
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("JWT_INVALID", "invalid token format"))
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("JWT_INVALID", "empty token"))
			return
		}

		// Validate the JWT.
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("JWT_EXPIRED", "jwt expired"))
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("JWT_INVALID", "jwt invalid"))
			return
		}

		// Set the authenticated user in the context.
		user := &domain.AuthenticatedUser{
			PlayerID: claims.PlayerID,
			TgUserID: claims.TgUserID,
			Role:     claims.Role,
		}
		c.Set(string(authenticatedUserKey), user)

		c.Next()
	}
}

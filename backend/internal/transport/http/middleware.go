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

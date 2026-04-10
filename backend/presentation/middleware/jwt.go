package middleware

import (
	"poker-club-backend/infrastructure/services"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates JWT access tokens and sets user_id in context
func JWTAuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Expect format: Bearer <token>
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		// Validate access token
		userID, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			if err == services.ErrTokenExpired {
				c.JSON(401, gin.H{"error": "token expired"})
				c.Abort()
				return
			}
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user_id in context
		c.Set("user_id", userID)
		c.Next()
	}
}

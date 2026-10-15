package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/demo"
	"poker-club/backend/internal/domain"
)

const demoSessionKey = "demo_session"

// DemoAuthMiddleware validates an opaque demo bearer token against the demo DB.
// It is independent from production JWT AuthMiddleware.
func DemoAuthMiddleware(demoSvc *demo.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if demoSvc == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, errorResponse("DEMO_UNAVAILABLE", "demo mode is unavailable"))
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("AUTHENTICATION_REQUIRED", "authentication required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("DEMO_TOKEN_INVALID", "invalid demo token format"))
			return
		}

		session, player, err := demoSvc.Authenticate(c.Request.Context(), parts[1])
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("DEMO_TOKEN_INVALID", "demo session invalid or expired"))
				return
			}
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, errorResponse("DEMO_UNAVAILABLE", "demo mode is unavailable"))
			return
		}

		user := &domain.AuthenticatedUser{
			PlayerID: player.ID,
			TgUserID: player.TgUserID,
			Role:     "owner",
		}
		c.Set(string(authenticatedUserKey), user)
		c.Set(demoSessionKey, session)
		c.Next()
	}
}

package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/demo"
	"poker-club/backend/internal/service"
)

// DemoHandler handles demo session lifecycle endpoints.
type DemoHandler struct {
	demoSvc *demo.Service
}

// NewDemoHandler creates a DemoHandler. demoSvc may be nil (demo disabled).
func NewDemoHandler(demoSvc *demo.Service) *DemoHandler {
	return &DemoHandler{demoSvc: demoSvc}
}

// CreateSession handles POST /api/v1/demo/sessions.
func (h *DemoHandler) CreateSession(c *gin.Context) {
	if h.demoSvc == nil {
		writeError(c, http.StatusServiceUnavailable, "DEMO_UNAVAILABLE", "demo mode is unavailable")
		return
	}

	result, err := h.demoSvc.CreateSession(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "DEMO_UNAVAILABLE", "demo mode is unavailable")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"access_token": result.AccessToken,
		"club_id":      result.ClubID,
		"expires_at":   result.ExpiresAt.UTC().Format(time.RFC3339),
		"token_type":   "Bearer",
	})
}

// Me handles GET /api/v1/demo/me for the demo visitor (club owner).
func (h *DemoHandler) Me(c *gin.Context, svc *service.Service) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}
	player, err := svc.GetPlayerByTgUserID(c.Request.Context(), tgUserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            player.ID,
		"first_name":    player.FirstName,
		"last_name":     player.LastName,
		"nickname":      player.Nickname,
		"tg_user_name":  player.TgUserName,
		"email":         player.Email,
		"phone_number":  player.PhoneNumber,
		"tg_user_id":    player.TgUserID,
		"has_password":  player.HasPassword(),
		"created_at":    player.CreatedAt,
		"updated_at":    player.UpdatedAt,
		"demo":          true,
	})
}

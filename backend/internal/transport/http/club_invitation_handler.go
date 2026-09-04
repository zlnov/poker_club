package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/service"
)

// ClubInvitationHandler handles HTTP requests for club invitation management.
// Endpoints:
//   - GET  /api/v1/clubs/{clubId}/invites — list active invitations
//   - POST /api/v1/clubs/{clubId}/invites — create an invitation
type ClubInvitationHandler struct {
	svc *service.Service
}

// NewClubInvitationHandler creates a new ClubInvitationHandler.
func NewClubInvitationHandler(svc *service.Service) *ClubInvitationHandler {
	return &ClubInvitationHandler{svc: svc}
}

// createInvitationRequest represents the request body for creating an invitation.
type createInvitationRequest struct {
	TgUserID int64 `json:"tg_user_id" binding:"required"`
}

// ListInvites handles GET /api/v1/clubs/{clubId}/invites.
// Returns all pending invitations (members with status='pending' and accepted=false).
func (h *ClubInvitationHandler) ListInvites(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	clubID, err := parseIDParam(c, "clubId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	members, err := h.svc.GetClubMembers(c.Request.Context(), tgUserID, clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	// Filter for pending invitations (status='pending', accepted=false).
	var invites []gin.H
	for _, m := range members {
		if m.Status == "pending" && !m.Accepted {
			invites = append(invites, gin.H{
				"player_id":   m.PlayerID,
				"status":      m.Status,
				"accepted":    m.Accepted,
				"created_at":  m.CreatedAt,
				"updated_at":  m.UpdatedAt,
				"player":      serializePlayer(&m.Player),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invites": invites,
	})
}

// CreateInvite handles POST /api/v1/clubs/{clubId}/invites.
// Creates an invitation for a player to join the club.
func (h *ClubInvitationHandler) CreateInvite(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	clubID, err := parseIDParam(c, "clubId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid club ID")
		return
	}

	var req createInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	player, club, err := h.svc.InviteMember(c.Request.Context(), tgUserID, clubID, req.TgUserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"player": serializePlayer(player),
		"club":   serializeClub(club),
		"message": "invitation created",
	})
}

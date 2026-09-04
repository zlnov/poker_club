package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/service"
)

// MembershipRequestHandler handles HTTP requests for membership request management.
// Endpoints:
//   - GET  /api/v1/clubs/{clubId}/member-requests                     — list pending requests
//   - POST /api/v1/clubs/{clubId}/members/{playerId}/approve          — approve a member
//   - POST /api/v1/clubs/{clubId}/members/{playerId}/reject           — reject a member
type MembershipRequestHandler struct {
	svc *service.Service
}

// NewMembershipRequestHandler creates a new MembershipRequestHandler.
func NewMembershipRequestHandler(svc *service.Service) *MembershipRequestHandler {
	return &MembershipRequestHandler{svc: svc}
}

// ListMemberRequests handles GET /api/v1/clubs/{clubId}/member-requests.
// Returns all members with pending status (membership requests awaiting confirmation).
func (h *MembershipRequestHandler) ListMemberRequests(c *gin.Context) {
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

	// Filter for pending members (status='pending').
	// Members with status='pending' and accepted=true are awaiting confirmation.
	// Members with status='pending' and accepted=false are awaiting response.
	var requests []gin.H
	for _, m := range members {
		if m.Status == "pending" {
			requests = append(requests, gin.H{
				"player_id":   m.PlayerID,
				"role":        m.Role,
				"status":      m.Status,
				"accepted":    m.Accepted,
				"created_at":  m.CreatedAt,
				"updated_at":  m.UpdatedAt,
				"player":      serializePlayer(&m.Player),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
	})
}

// ApproveMember handles POST /api/v1/clubs/{clubId}/members/{playerId}/approve.
// Confirms a member's entry into the club (pending → active).
func (h *MembershipRequestHandler) ApproveMember(c *gin.Context) {
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

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	player, club, err := h.svc.ConfirmEntry(c.Request.Context(), tgUserID, clubID, playerID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player": serializePlayer(player),
		"club":   serializeClub(club),
		"message": "member entry confirmed",
	})
}

// RejectMember handles POST /api/v1/clubs/{clubId}/members/{playerId}/reject.
// Rejects a member's entry into the club (sets status to 'left').
func (h *MembershipRequestHandler) RejectMember(c *gin.Context) {
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

	playerID, err := parseIDParam(c, "playerId")
	if err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid player ID")
		return
	}

	// Use ChangeMemberStatus to set the member's status to 'left' (rejection).
	player, err := h.svc.ChangeMemberStatus(c.Request.Context(), tgUserID, clubID, playerID, "left")
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player": serializePlayer(player),
		"message": "member rejected",
	})
}

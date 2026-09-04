package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/service"
)

// ClubMemberHandler handles HTTP requests for club member management.
// Endpoints:
//   - GET    /api/v1/clubs/{clubId}/members              — list all members
//   - POST   /api/v1/clubs/{clubId}/members              — invite a member
//   - PATCH  /api/v1/clubs/{clubId}/members/{playerId}   — update member role/status
//   - DELETE /api/v1/clubs/{clubId}/members/{playerId}   — remove member
type ClubMemberHandler struct {
	svc *service.Service
}

// NewClubMemberHandler creates a new ClubMemberHandler.
func NewClubMemberHandler(svc *service.Service) *ClubMemberHandler {
	return &ClubMemberHandler{svc: svc}
}

// inviteMemberRequest represents the request body for inviting a member.
type inviteMemberRequest struct {
	TgUserID int64 `json:"tg_user_id" binding:"required"`
}

// updateMemberRequest represents the request body for updating a member.
type updateMemberRequest struct {
	Role   *string `json:"role,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ListMembers handles GET /api/v1/clubs/{clubId}/members.
// Returns all members of the club with their player info.
func (h *ClubMemberHandler) ListMembers(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"members": serializeClubMembers(members),
	})
}

// InviteMember handles POST /api/v1/clubs/{clubId}/members.
// Invites a player to the club by their Telegram user ID.
func (h *ClubMemberHandler) InviteMember(c *gin.Context) {
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

	var req inviteMemberRequest
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
	})
}

// UpdateMember handles PATCH /api/v1/clubs/{clubId}/members/{playerId}.
// Updates a member's role or status.
func (h *ClubMemberHandler) UpdateMember(c *gin.Context) {
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

	var req updateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Handle role changes.
	if req.Role != nil {
		role := *req.Role
		switch role {
		case "admin":
			player, err := h.svc.AssignAdmin(c.Request.Context(), tgUserID, clubID, playerID)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"player": serializePlayer(player),
				"message": "admin role assigned",
			})
			return
		case "member":
			player, err := h.svc.RemoveAdmin(c.Request.Context(), tgUserID, clubID, playerID)
			if err != nil {
				writeServiceError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"player": serializePlayer(player),
				"message": "admin role removed",
			})
			return
		default:
			writeError(c, http.StatusBadRequest, "INVALID_MEMBER_ROLE", "invalid role: "+role)
			return
		}
	}

	// Handle status changes.
	if req.Status != nil {
		player, err := h.svc.ChangeMemberStatus(c.Request.Context(), tgUserID, clubID, playerID, *req.Status)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"player": serializePlayer(player),
			"message": "member status updated",
		})
		return
	}

	writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "no fields to update")
}

// RemoveMember handles DELETE /api/v1/clubs/{clubId}/members/{playerId}.
// Removes a member from the club by setting their status to 'left'.
func (h *ClubMemberHandler) RemoveMember(c *gin.Context) {
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

	player, err := h.svc.RemoveMember(c.Request.Context(), tgUserID, clubID, playerID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player": serializePlayer(player),
		"message": "member removed",
	})
}

// serializeClubMember converts a domain.ClubMemberWithPlayer to a JSON-serializable map.
func serializeClubMember(m *domain.ClubMemberWithPlayer) gin.H {
	result := gin.H{
		"id":         m.ID,
		"club_id":    m.ClubID,
		"player_id":  m.PlayerID,
		"role":       m.Role,
		"status":     m.Status,
		"accepted":   m.Accepted,
		"created_at": m.CreatedAt,
		"updated_at": m.UpdatedAt,
		"player":     serializePlayer(&m.Player),
	}
	return result
}

// serializeClubMembers converts a list of domain.ClubMemberWithPlayer to a JSON-serializable slice.
func serializeClubMembers(members []*domain.ClubMemberWithPlayer) []gin.H {
	result := make([]gin.H, len(members))
	for i, m := range members {
		result[i] = serializeClubMember(m)
	}
	return result
}

// serializePlayer converts a domain.Player to a JSON-serializable map.
func serializePlayer(p *domain.Player) gin.H {
	result := gin.H{
		"id":           p.ID,
		"first_name":   p.FirstName,
		"last_name":    p.LastName,
		"nickname":     p.Nickname,
		"created_at":   p.CreatedAt,
		"updated_at":   p.UpdatedAt,
	}
	if p.PhoneNumber != nil {
		result["phone_number"] = *p.PhoneNumber
	}
	if p.Email != nil {
		result["email"] = *p.Email
	}
	if p.TgUserID != nil {
		result["tg_user_id"] = *p.TgUserID
	}
	return result
}

package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"poker-club/backend/internal/domain"
	"poker-club/backend/internal/service"
)

// ClubHandler handles HTTP requests for club management.
// Endpoints:
//   - GET    /api/v1/clubs           — list clubs for the authenticated user
//   - POST   /api/v1/clubs           — create a new club
//   - GET    /api/v1/clubs/{clubId}  — get club information
//   - PATCH  /api/v1/clubs/{clubId}  — update club (change name)
type ClubHandler struct {
	svc *service.Service
}

// NewClubHandler creates a new ClubHandler.
func NewClubHandler(svc *service.Service) *ClubHandler {
	return &ClubHandler{svc: svc}
}

// createClubRequest represents the request body for creating a club.
type createClubRequest struct {
	Name string `json:"name" binding:"required"`
}

// updateClubRequest represents the request body for updating a club.
type updateClubRequest struct {
	Name string `json:"name,omitempty"`
}

// ListClubs handles GET /api/v1/clubs.
// Returns all clubs where the authenticated user is a member (any role).
func (h *ClubHandler) ListClubs(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	clubs, err := h.svc.GetUserClubsAll(c.Request.Context(), tgUserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clubs": serializeClubs(clubs),
	})
}

// CreateClub handles POST /api/v1/clubs.
// Creates a new club with the authenticated user as owner.
func (h *ClubHandler) CreateClub(c *gin.Context) {
	tgUserID, ok := getTgUserIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
		return
	}

	var req createClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	// Get the player to obtain first_name, last_name, nickname for club creation.
	player, err := h.svc.GetPlayerByTgUserID(c.Request.Context(), tgUserID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	firstName := player.FirstName
	lastName := player.LastName
	nickname := player.Nickname

	club, err := h.svc.CreateClub(c.Request.Context(), tgUserID, firstName, lastName, nickname, req.Name)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, serializeClub(club))
}

// GetClub handles GET /api/v1/clubs/{clubId}.
// Returns information about a specific club.
func (h *ClubHandler) GetClub(c *gin.Context) {
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

	// Check permission: user must be a member of the club.
	if err := h.svc.CheckPermission(c.Request.Context(), tgUserID, clubID, service.PermViewClub); err != nil {
		writeServiceError(c, err)
		return
	}

	club, err := h.svc.GetClubInfo(c.Request.Context(), clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	// Get the authenticated user's role in this club.
	member, memberErr := h.svc.GetClubMember(c.Request.Context(), clubID, tgUserID)

	response := serializeClub(club)
	if memberErr == nil {
		response["role"] = member.Role
		response["is_owner"] = member.Role == "owner"
		response["is_admin"] = member.Role == "admin"
	}

	c.JSON(http.StatusOK, response)
}

// UpdateClub handles PATCH /api/v1/clubs/{clubId}.
// Updates club parameters (currently only name).
func (h *ClubHandler) UpdateClub(c *gin.Context) {
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

	var req updateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request: "+err.Error())
		return
	}

	if req.Name != "" {
		if err := h.svc.ChangeClubName(c.Request.Context(), tgUserID, clubID, req.Name); err != nil {
			writeServiceError(c, err)
			return
		}
	}

	// Return updated club info.
	club, err := h.svc.GetClubInfo(c.Request.Context(), clubID)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, serializeClub(club))
}

// CloseClub handles DELETE /api/v1/clubs/{clubId}.
// Closes (deletes) a club. Only the owner can perform this action.
func (h *ClubHandler) CloseClub(c *gin.Context) {
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

	if err := h.svc.CloseClub(c.Request.Context(), tgUserID, clubID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "club closed successfully",
	})
}

// serializeClub converts a domain.Club to a JSON-serializable map.
func serializeClub(club *domain.Club) gin.H {
	result := gin.H{
		"id":         club.ID,
		"name":       club.Name,
		"created_at": club.CreatedAt,
		"updated_at": club.UpdatedAt,
	}
	if club.TgChatID != nil {
		result["tg_chat_id"] = *club.TgChatID
	}
	return result
}

// serializeClubs converts a list of domain.Club to a JSON-serializable slice.
func serializeClubs(clubs []*domain.Club) []gin.H {
	result := make([]gin.H, len(clubs))
	for i, club := range clubs {
		result[i] = serializeClub(club)
	}
	return result
}

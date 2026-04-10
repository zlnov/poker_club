package handlers

import (
	"net/http"

	"poker-club-backend/application/dtos"
	"poker-club-backend/application/usecases"

	"github.com/gin-gonic/gin"
)

// ClubHandler handles club-related HTTP requests
type ClubHandler struct {
	clubUseCase *usecases.ClubUseCase
}

// NewClubHandler creates a new ClubHandler
func NewClubHandler(clubUseCase *usecases.ClubUseCase) *ClubHandler {
	return &ClubHandler{
		clubUseCase: clubUseCase,
	}
}

// CreateClub handles POST /clubs
func (h *ClubHandler) CreateClub(c *gin.Context) {
	var req dtos.CreateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get creator ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	req.CreatorID = userID.(int64)

	response, err := h.clubUseCase.CreateClub(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetClubMembers handles GET /clubs/:id/members
func (h *ClubHandler) GetClubMembers(c *gin.Context) {
	clubID := c.GetInt64("club_id")
	if clubID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club id"})
		return
	}

	members, err := h.clubUseCase.GetClubMembers(c.Request.Context(), clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"club_id": clubID,
		"members": members,
	})
}

// ApproveMember handles POST /clubs/:id/members/approve
func (h *ClubHandler) ApproveMember(c *gin.Context) {
	clubID := c.GetInt64("club_id")
	if clubID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club id"})
		return
	}

	var req dtos.ApproveMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ClubID = clubID
	req.ApproverID = c.GetInt64("user_id")

	err := h.clubUseCase.ApproveMember(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member approved"})
}

// RejectMember handles POST /clubs/:id/members/reject
func (h *ClubHandler) RejectMember(c *gin.Context) {
	clubID := c.GetInt64("club_id")
	if clubID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid club id"})
		return
	}

	var req dtos.RejectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ClubID = clubID
	req.RejecterID = c.GetInt64("user_id")

	err := h.clubUseCase.RejectMember(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member rejected"})
}

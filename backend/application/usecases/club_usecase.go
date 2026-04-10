package usecases

import (
	"context"

	"poker-club-backend/application/dtos"
	"poker-club-backend/domain"
)

// ClubUseCase handles club-related use cases
type ClubUseCase struct {
	clubService *domain.ClubService
}

// NewClubUseCase creates a new ClubUseCase
func NewClubUseCase(clubService *domain.ClubService) *ClubUseCase {
	return &ClubUseCase{
		clubService: clubService,
	}
}

// CreateClub creates a new club
func (uc *ClubUseCase) CreateClub(ctx context.Context, req dtos.CreateClubRequest) (*dtos.CreateClubResponse, error) {
	club, err := uc.clubService.CreateClub(ctx, req.Name, req.CreatorID)
	if err != nil {
		return nil, err
	}

	return &dtos.CreateClubResponse{
		ID:        club.ID,
		Name:      club.Name,
		CreatedAt: club.CreatedAt,
	}, nil
}

// ApproveMember approves a club member
func (uc *ClubUseCase) ApproveMember(ctx context.Context, req dtos.ApproveMemberRequest) error {
	return uc.clubService.ApproveMember(ctx, req.ClubID, req.MemberID, req.ApproverID)
}

// RejectMember rejects a club member
func (uc *ClubUseCase) RejectMember(ctx context.Context, req dtos.RejectMemberRequest) error {
	return uc.clubService.RejectMember(ctx, req.ClubID, req.MemberID, req.RejecterID)
}

// GetClubMembers returns all active members of a club
func (uc *ClubUseCase) GetClubMembers(ctx context.Context, clubID int64) ([]dtos.ClubMemberDTO, error) {
	members, err := uc.clubService.GetClubMembers(ctx, clubID)
	if err != nil {
		return nil, err
	}

	result := make([]dtos.ClubMemberDTO, 0, len(members))
	for _, member := range members {
		dto := dtos.ClubMemberDTO{
			ID:       member.ID,
			ClubID:   member.ClubID,
			PlayerID: member.PlayerID,
			Role:     domain.Role(member.Role),
			Status:   domain.MemberStatus(member.Status),
		}
		result = append(result, dto)
	}

	return result, nil
}

// CheckMemberAccess checks if a player has access to a club
func (uc *ClubUseCase) CheckMemberAccess(ctx context.Context, clubID, playerID int64) (bool, error) {
	return uc.clubService.CheckMemberAccess(ctx, clubID, playerID)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"poker-club/backend/internal/domain"
)

// Permission represents an action that can be performed on a club.
type Permission string

const (
	PermViewClub  Permission = "view_club"
	PermEditClub  Permission = "edit_club"
	PermCloseClub Permission = "close_club"
)

// rolePermissions maps each role to the permissions it grants.
var rolePermissions = map[string][]Permission{
	"owner":  {PermViewClub, PermEditClub, PermCloseClub},
	"admin":  {PermViewClub},
	"member": {},
}

// Service provides business logic operations.
// It depends on repositories for data persistence.
type Service struct {
	repos *domain.Repositories
	log   *slog.Logger
}

// New creates a new Service instance.
func New(repos *domain.Repositories, log *slog.Logger) *Service {
	return &Service{
		repos: repos,
		log:   log,
	}
}

// HealthCheck verifies that all dependencies are accessible.
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.repos.Clubs.Ping(ctx)
}

// RegisterTelegramUser finds or creates a player from Telegram user data.
// If the player already exists (matched by tg_user_id), it is returned.
// Otherwise a new player is created with the provided Telegram data.
func (s *Service) RegisterTelegramUser(ctx context.Context, tgUserID int64, firstName, lastName, nickname string) (*domain.Player, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err == nil {
		_ = s.repos.Players.UpdateLastSeen(ctx, player.ID)
		return player, nil
	}

	player = &domain.Player{
		FirstName:   firstName,
		LastName:    lastName,
		Nickname:    nickname,
		PhoneNumber: "",
		Email:       "",
		Password:    "",
		TgUserID:    &tgUserID,
	}
	playerID, err := s.repos.Players.Create(ctx, player)
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}
	player.ID = playerID
	return player, nil
}

// CreateClub creates a new club with the given Telegram user as owner.
// If the player does not exist yet, it is created from the provided Telegram data.
func (s *Service) CreateClub(ctx context.Context, tgUserID int64, firstName, lastName, nickname, clubName string) (*domain.Club, error) {
	player, err := s.RegisterTelegramUser(ctx, tgUserID, firstName, lastName, nickname)
	if err != nil {
		return nil, err
	}

	club := &domain.Club{
		Name: clubName,
	}
	clubID, err := s.repos.Clubs.Create(ctx, club)
	if err != nil {
		return nil, fmt.Errorf("failed to create club: %w", err)
	}
	club.ID = clubID

	// Create club member with owner role.
	member := &domain.ClubMember{
		ClubID:   clubID,
		PlayerID: player.ID,
		Role:     "owner",
		Status:   "active",
	}
	if _, err := s.repos.ClubMembers.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to create club member: %w", err)
	}

	s.log.Info("club created",
		"club_id", clubID,
		"club_name", clubName,
		"owner_tg_user_id", tgUserID,
	)

	return club, nil
}

// GetClubInfo returns information about a club.
func (s *Service) GetClubInfo(ctx context.Context, clubID int64) (*domain.Club, error) {
	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}
	return club, nil
}

// GetUserClubs returns all clubs where the given Telegram user is the owner.
func (s *Service) GetUserClubs(ctx context.Context, tgUserID int64) ([]*domain.Club, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, err
	}
	return s.repos.Clubs.GetByOwner(ctx, player.ID)
}

// BindGroupToClub binds a Telegram group chat to a club.
// The user must be the owner of the club, the club must not already have a
// tg_chat_id, and the chat_id must not be used by another club.
func (s *Service) BindGroupToClub(ctx context.Context, tgUserID int64, clubID int64, tgChatID int64) (*domain.Club, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermEditClub); err != nil {
		return nil, err
	}

	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	if club.TgChatID != nil {
		return nil, errors.New("club already has a bound group")
	}

	existing, err := s.repos.Clubs.GetByTgChatID(ctx, tgChatID)
	if err == nil && existing != nil {
		return nil, errors.New("this group is already bound to another club")
	}

	if err := s.repos.Clubs.BindTgChatID(ctx, clubID, tgChatID); err != nil {
		return nil, fmt.Errorf("failed to bind group: %w", err)
	}

	s.log.Info("group bound to club",
		"club_id", clubID,
		"tg_chat_id", tgChatID,
		"tg_user_id", tgUserID,
	)

	club.TgChatID = &tgChatID
	return club, nil
}

// ChangeClubName changes the name of a club. Only the owner can perform this action.
func (s *Service) ChangeClubName(ctx context.Context, tgUserID int64, clubID int64, newClubName string) error {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermEditClub); err != nil {
		return err
	}

	if err := s.repos.Clubs.UpdateName(ctx, clubID, newClubName); err != nil {
		return fmt.Errorf("failed to update club name: %w", err)
	}

	s.log.Info("club name changed",
		"club_id", clubID,
		"new_name", newClubName,
		"tg_user_id", tgUserID,
	)
	return nil
}

// CloseClub deletes a club and all related data. Only the owner can perform this action.
func (s *Service) CloseClub(ctx context.Context, tgUserID int64, clubID int64) error {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermCloseClub); err != nil {
		return err
	}

	if err := s.repos.Clubs.Delete(ctx, clubID); err != nil {
		return fmt.Errorf("failed to delete club: %w", err)
	}

	s.log.Info("club closed",
		"club_id", clubID,
		"tg_user_id", tgUserID,
	)
	return nil
}

// checkPermission verifies that the Telegram user has the required permission
// for the given club. It resolves tg_user_id -> player -> club_member -> role.
func (s *Service) checkPermission(ctx context.Context, tgUserID int64, clubID int64, perm Permission) error {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return fmt.Errorf("access denied: user not found: %w", err)
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return fmt.Errorf("access denied: user is not a member of this club: %w", err)
	}

	allowed := false
	for _, p := range rolePermissions[member.Role] {
		if p == perm {
			allowed = true
			break
		}
	}

	if !allowed {
		s.log.Warn("access denied",
			"tg_user_id", tgUserID,
			"club_id", clubID,
			"role", member.Role,
			"required_permission", perm,
		)
		return errors.New("access denied: insufficient permissions")
	}

	return nil
}

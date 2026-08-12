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
	PermViewClub           Permission = "view_club"
	PermEditClub           Permission = "edit_club"
	PermCloseClub          Permission = "close_club"
	PermInviteMember       Permission = "invite_member"
	PermListMembers        Permission = "list_members"
	PermRemoveMember       Permission = "remove_member"
	PermChangeMemberStatus Permission = "change_member_status"
	PermAssignAdmin        Permission = "assign_admin"
	PermRemoveAdmin        Permission = "remove_admin"
	PermConfirmEntry       Permission = "confirm_entry"
)

// rolePermissions maps each role to the permissions it grants.
var rolePermissions = map[string][]Permission{
	"owner": {
		PermViewClub, PermEditClub, PermCloseClub,
		PermInviteMember, PermListMembers, PermRemoveMember,
		PermChangeMemberStatus, PermAssignAdmin, PermRemoveAdmin, PermConfirmEntry,
	},
	"admin": {
		PermViewClub,
		PermInviteMember, PermListMembers, PermRemoveMember,
		PermChangeMemberStatus, PermConfirmEntry,
	},
	"member": {
		PermListMembers,
	},
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

// InviteMember invites a player to a club by their Telegram user ID.
// The inviter must be the owner or an admin. The invited player must already
// have a player record (i.e. have been registered via NewChatMembers). A
// club_member with status 'pending' and accepted=false is created.
func (s *Service) InviteMember(ctx context.Context, tgUserID int64, clubID int64, inviteeTgUserID int64) (*domain.Player, *domain.Club, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermInviteMember); err != nil {
		return nil, nil, err
	}

	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, nil, err
	}

	player, err := s.repos.Players.GetByTgUserID(ctx, inviteeTgUserID)
	if err != nil {
		return nil, nil, errors.New("пользователь с таким username не найден")
	}

	// Check if the player is already a member of this club.
	existing, _ := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if existing != nil {
		return nil, nil, errors.New("пользователь уже является участником клуба")
	}

	member := &domain.ClubMember{
		ClubID:   clubID,
		PlayerID: player.ID,
		Role:     "member",
		Status:   "pending",
		Accepted: false,
	}
	if _, err := s.repos.ClubMembers.Create(ctx, member); err != nil {
		return nil, nil, fmt.Errorf("failed to create club member: %w", err)
	}

	s.log.Info("member invited",
		"club_id", clubID,
		"player_id", player.ID,
		"inviter_tg_user_id", tgUserID,
		"invitee_tg_user_id", inviteeTgUserID,
	)

	return player, club, nil
}

// GetClubByTgChatID returns the club bound to the given Telegram chat ID.
func (s *Service) GetClubByTgChatID(ctx context.Context, tgChatID int64) (*domain.Club, error) {
	return s.repos.Clubs.GetByTgChatID(ctx, tgChatID)
}

// GetPlayerByUsername returns a player by their Telegram username (nickname).
func (s *Service) GetPlayerByUsername(ctx context.Context, username string) (*domain.Player, error) {
	return s.repos.Players.GetByNickname(ctx, username)
}

// GetPlayerByTgUserID returns a player by their Telegram user ID.
func (s *Service) GetPlayerByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	return s.repos.Players.GetByTgUserID(ctx, tgUserID)
}

// GetClubMember returns the club membership record for a given player and club.
func (s *Service) GetClubMember(ctx context.Context, clubID, playerID int64) (*domain.ClubMember, error) {
	return s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
}

// GetUserRole returns the role of a Telegram user in a specific club.
func (s *Service) GetUserRole(ctx context.Context, tgUserID int64, clubID int64) (string, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return "", err
	}
	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

// AcceptInvitation allows an invited user to accept an invitation to a club.
// The user must have a pending invitation (status='pending', accepted=false).
// After accepting, accepted is set to true and the invitation awaits
// confirmation from the owner or admin.
func (s *Service) AcceptInvitation(ctx context.Context, tgUserID int64, clubID int64) (*domain.Player, *domain.Club, []int64, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return nil, nil, nil, errors.New("у вас нет приглашения в этот клуб")
	}

	if member.Status != "pending" || member.Accepted {
		return nil, nil, nil, errors.New("приглашение уже обработано")
	}

	if err := s.repos.ClubMembers.UpdateAccepted(ctx, clubID, player.ID, true); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get owner and admin tg_user_ids for notification.
	notifyIDs, err := s.getOwnerAndAdminTgUserIDs(ctx, clubID)
	if err != nil {
		return nil, nil, nil, err
	}

	s.log.Info("invitation accepted",
		"club_id", clubID,
		"player_id", player.ID,
		"tg_user_id", tgUserID,
	)

	return player, club, notifyIDs, nil
}

// RejectInvitation allows an invited user to reject an invitation to a club.
// The user must have a pending invitation (status='pending', accepted=false).
// After rejecting, the member status is set to 'left'.
func (s *Service) RejectInvitation(ctx context.Context, tgUserID int64, clubID int64) (*domain.Club, []int64, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return nil, nil, errors.New("у вас нет приглашения в этот клуб")
	}

	if member.Status != "pending" || member.Accepted {
		return nil, nil, errors.New("приглашение уже обработано")
	}

	if err := s.repos.ClubMembers.UpdateStatus(ctx, clubID, player.ID, "left"); err != nil {
		return nil, nil, fmt.Errorf("failed to reject invitation: %w", err)
	}

	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, nil, err
	}

	// Get owner and admin tg_user_ids for notification.
	notifyIDs, err := s.getOwnerAndAdminTgUserIDs(ctx, clubID)
	if err != nil {
		return nil, nil, err
	}

	s.log.Info("invitation rejected",
		"club_id", clubID,
		"player_id", player.ID,
		"tg_user_id", tgUserID,
	)

	return club, notifyIDs, nil
}

// ConfirmEntry allows the owner or admin to confirm a user's entry into the club.
// The user must have accepted the invitation (accepted=true, status='pending').
// After confirmation, the member status is set to 'active'.
func (s *Service) ConfirmEntry(ctx context.Context, tgUserID int64, clubID int64, playerID int64) (*domain.Player, *domain.Club, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermConfirmEntry); err != nil {
		return nil, nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return nil, nil, errors.New("участник не найден в клубе")
	}

	if member.Status != "pending" || !member.Accepted {
		return nil, nil, errors.New("приглашение не ожидает подтверждения")
	}

	if err := s.repos.ClubMembers.UpdateStatus(ctx, clubID, playerID, "active"); err != nil {
		return nil, nil, fmt.Errorf("failed to confirm entry: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, nil, err
	}

	club, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, nil, err
	}

	s.log.Info("member entry confirmed",
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, club, nil
}

// GetClubMembers returns all members of a club with their player info.
// The requesting user must have at least the list_members permission.
func (s *Service) GetClubMembers(ctx context.Context, tgUserID int64, clubID int64) ([]*domain.ClubMemberWithPlayer, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermListMembers); err != nil {
		return nil, err
	}

	members, err := s.repos.ClubMembers.GetByClubWithPlayers(ctx, clubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get club members: %w", err)
	}

	return members, nil
}

// RemoveMember removes a member from the club by setting their status to 'left'.
// The requesting user must be the owner or an admin. The owner cannot be removed.
func (s *Service) RemoveMember(ctx context.Context, tgUserID int64, clubID int64, playerID int64) (*domain.Player, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermRemoveMember); err != nil {
		return nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return nil, errors.New("участник не найден в клубе")
	}

	if member.Role == "owner" {
		return nil, errors.New("нельзя исключить владельца клуба")
	}

	if err := s.repos.ClubMembers.UpdateStatus(ctx, clubID, playerID, "left"); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	s.log.Info("member removed",
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, nil
}

// ChangeMemberStatus changes a member's status (e.g., active → banned, banned → active).
// The requesting user must be the owner or an admin. The owner's status cannot be changed.
func (s *Service) ChangeMemberStatus(ctx context.Context, tgUserID int64, clubID int64, playerID int64, status string) (*domain.Player, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermChangeMemberStatus); err != nil {
		return nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return nil, errors.New("участник не найден в клубе")
	}

	if member.Role == "owner" {
		return nil, errors.New("нельзя изменить статус владельца клуба")
	}

	if err := s.repos.ClubMembers.UpdateStatus(ctx, clubID, playerID, status); err != nil {
		return nil, fmt.Errorf("failed to change member status: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	s.log.Info("member status changed",
		"club_id", clubID,
		"player_id", playerID,
		"new_status", status,
		"tg_user_id", tgUserID,
	)

	return player, nil
}

// AssignAdmin assigns the admin role to a club member.
// Only the owner can perform this action. The owner cannot be assigned as admin.
func (s *Service) AssignAdmin(ctx context.Context, tgUserID int64, clubID int64, playerID int64) (*domain.Player, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermAssignAdmin); err != nil {
		return nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return nil, errors.New("участник не найден в клубе")
	}

	if member.Role == "owner" {
		return nil, errors.New("владелец не может быть назначен администратором")
	}

	if err := s.repos.ClubMembers.UpdateRole(ctx, clubID, playerID, "admin"); err != nil {
		return nil, fmt.Errorf("failed to assign admin: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	s.log.Info("admin assigned",
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, nil
}

// RemoveAdmin removes the admin role from a club member, setting it back to 'member'.
// Only the owner can perform this action.
func (s *Service) RemoveAdmin(ctx context.Context, tgUserID int64, clubID int64, playerID int64) (*domain.Player, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermRemoveAdmin); err != nil {
		return nil, err
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return nil, errors.New("участник не найден в клубе")
	}

	if member.Role != "admin" {
		return nil, errors.New("участник не является администратором")
	}

	if err := s.repos.ClubMembers.UpdateRole(ctx, clubID, playerID, "member"); err != nil {
		return nil, fmt.Errorf("failed to remove admin: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	s.log.Info("admin removed",
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, nil
}

// getOwnerAndAdminTgUserIDs returns the Telegram user IDs of the owner and all
// admins of the given club. Used for sending notifications.
func (s *Service) getOwnerAndAdminTgUserIDs(ctx context.Context, clubID int64) ([]int64, error) {
	members, err := s.repos.ClubMembers.GetByClubWithPlayers(ctx, clubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get club members: %w", err)
	}

	var ids []int64
	for _, m := range members {
		if m.Role == "owner" || m.Role == "admin" {
			if m.Player.TgUserID != nil {
				ids = append(ids, *m.Player.TgUserID)
			}
		}
	}
	return ids, nil
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

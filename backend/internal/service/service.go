package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"poker-club/backend/internal/domain"
)

// Permission represents an action that can be performed on a club.
type Permission string

const (
	PermViewClub               Permission = "view_club"
	PermEditClub               Permission = "edit_club"
	PermCloseClub              Permission = "close_club"
	PermInviteMember           Permission = "invite_member"
	PermListMembers            Permission = "list_members"
	PermRemoveMember           Permission = "remove_member"
	PermChangeMemberStatus     Permission = "change_member_status"
	PermAssignAdmin            Permission = "assign_admin"
	PermRemoveAdmin            Permission = "remove_admin"
	PermConfirmEntry           Permission = "confirm_entry"
	PermCreateGame             Permission = "create_game"
	PermEditGame               Permission = "edit_game"
	PermCancelGame             Permission = "cancel_game"
	PermInviteToGame           Permission = "invite_to_game"
	PermManageGameParticipants Permission = "manage_game_participants"
)

// rolePermissions maps each role to the permissions it grants.
var rolePermissions = map[string][]Permission{
	"owner": {
		PermViewClub, PermEditClub, PermCloseClub,
		PermInviteMember, PermListMembers, PermRemoveMember,
		PermChangeMemberStatus, PermAssignAdmin, PermRemoveAdmin, PermConfirmEntry,
		PermCreateGame, PermEditGame, PermCancelGame,
		PermInviteToGame, PermManageGameParticipants,
	},
	"admin": {
		PermViewClub,
		PermInviteMember, PermListMembers, PermRemoveMember,
		PermChangeMemberStatus, PermConfirmEntry,
		PermCreateGame, PermEditGame,
		PermInviteToGame, PermManageGameParticipants,
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

// GetUserClubsAll returns all clubs where the given Telegram user is a member
// (any role: owner, admin, or member).
func (s *Service) GetUserClubsAll(ctx context.Context, tgUserID int64) ([]*domain.Club, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, err
	}
	return s.repos.Clubs.GetByPlayer(ctx, player.ID)
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

// --- Game management ---

// CreateGame creates a new game in the specified club and invites all active
// club members. The requesting user must have the create_game permission.
// The game and its initial game_participants records are created atomically
// in a single transaction.
func (s *Service) CreateGame(ctx context.Context, tgUserID int64, clubID int64, game *domain.Game) (*domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermCreateGame); err != nil {
		return nil, err
	}

	// Verify the club exists.
	_, err := s.repos.Clubs.GetByID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	// Check constraint 7.15: no active games in the club.
	_, err = s.repos.Games.GetActiveByClub(ctx, clubID)
	if err == nil {
		return nil, errors.New("в клубе уже есть активная игра")
	}

	// Verify the banker is a club member.
	bankerMember, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, game.BankerID)
	if err != nil {
		return nil, errors.New("банкир не является участником клуба")
	}
	game.BankerID = bankerMember.ID
	game.ClubID = clubID
	game.Status = "planned"

	// Get all active club members to invite.
	members, err := s.repos.ClubMembers.GetByClubWithPlayers(ctx, clubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get club members: %w", err)
	}

	var activeMembers []*domain.ClubMemberWithPlayer
	for _, m := range members {
		if m.Status == "active" {
			activeMembers = append(activeMembers, m)
		}
	}

	// Create the game and game_participants atomically.
	gameID, err := s.repos.Games.Create(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	game.ID = gameID

	for _, m := range activeMembers {
		participant := &domain.GameParticipant{
			GameID:   gameID,
			PlayerID: m.PlayerID,
			Status:   "invited",
		}
		if _, err := s.repos.GameParticipants.Create(ctx, participant); err != nil {
			return nil, fmt.Errorf("failed to create game participant: %w", err)
		}
	}

	s.log.Info("game created",
		"game_id", gameID,
		"club_id", clubID,
		"game_type", game.GameType,
		"created_by", tgUserID,
	)

	return game, nil
}

// GetGame returns a game by ID if the user has permission to view it.
func (s *Service) GetGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermViewClub); err != nil {
		return nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	return game, nil
}

// GetClubGames returns all games for a club.
func (s *Service) GetClubGames(ctx context.Context, tgUserID int64, clubID int64) ([]*domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermViewClub); err != nil {
		return nil, err
	}

	return s.repos.Games.GetByClub(ctx, clubID)
}

// UpdateGame updates game parameters. Only allowed in planned status.
// Only owner/admin can change parameters.
func (s *Service) UpdateGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64, game *domain.Game) (*domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermEditGame); err != nil {
		return nil, err
	}

	existing, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if existing.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	if existing.Status != "planned" {
		return nil, errors.New("изменение параметров доступно только для игр в статусе planned")
	}

	game.ID = existing.ID
	game.ClubID = existing.ClubID
	game.Status = existing.Status
	game.CreatedAt = existing.CreatedAt

	if err := s.repos.Games.Update(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	s.log.Info("game parameters updated",
		"game_id", gameID,
		"club_id", clubID,
		"tg_user_id", tgUserID,
	)

	return game, nil
}

// CancelGame cancels a game. Only owner/admin can cancel.
// Only allowed in planned status.
func (s *Service) CancelGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64) error {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermCancelGame); err != nil {
		return err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.ClubID != clubID {
		return errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return errors.New("отмена доступна только для игр в статусе planned")
	}

	if err := s.repos.Games.Cancel(ctx, gameID); err != nil {
		return fmt.Errorf("failed to cancel game: %w", err)
	}

	s.log.Info("game cancelled",
		"game_id", gameID,
		"club_id", clubID,
		"tg_user_id", tgUserID,
	)

	return nil
}

// InviteToGame invites a club member to a game. Only owner/admin can invite.
// The game must be in planned status.
func (s *Service) InviteToGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64, inviteeTgUserID int64) (*domain.Player, *domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermInviteToGame); err != nil {
		return nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return nil, nil, errors.New("приглашение доступно только для игр в статусе planned")
	}

	player, err := s.repos.Players.GetByTgUserID(ctx, inviteeTgUserID)
	if err != nil {
		return nil, nil, errors.New("пользователь не найден")
	}

	// Check if player is an active club member.
	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return nil, nil, errors.New("пользователь не является участником клуба")
	}
	if member.Status != "active" {
		return nil, nil, errors.New("пользователь не является активным участником клуба")
	}

	// Check if player is already a participant of this game.
	existing, _ := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if existing != nil {
		if existing.Status == "declined" {
			// Re-invite: update status to invited.
			existing.Status = "invited"
			if err := s.repos.GameParticipants.Update(ctx, existing); err != nil {
				return nil, nil, fmt.Errorf("failed to re-invite participant: %w", err)
			}
		} else {
			return nil, nil, errors.New("игрок уже является участником этой игры")
		}
	} else {
		participant := &domain.GameParticipant{
			GameID:   gameID,
			PlayerID: player.ID,
			Status:   "invited",
		}
		if _, err := s.repos.GameParticipants.Create(ctx, participant); err != nil {
			return nil, nil, fmt.Errorf("failed to create game participant: %w", err)
		}
	}

	s.log.Info("member invited to game",
		"game_id", gameID,
		"club_id", clubID,
		"player_id", player.ID,
		"inviter_tg_user_id", tgUserID,
	)

	return player, game, nil
}

// AcceptGameParticipation allows a player to accept a game invitation.
// The player must have an invited status for the game.
func (s *Service) AcceptGameParticipation(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Player, *domain.Game, []int64, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return nil, nil, nil, errors.New("принять участие можно только для игр в статусе planned")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return nil, nil, nil, errors.New("вы не приглашены в эту игру")
	}

	if participant.Status != "invited" {
		return nil, nil, nil, errors.New("приглашение уже обработано")
	}

	if err := s.repos.GameParticipants.UpdateStatus(ctx, gameID, player.ID, "accepted"); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to accept game participation: %w", err)
	}

	// Get owner and admin tg_user_ids for notification.
	notifyIDs, err := s.getOwnerAndAdminTgUserIDs(ctx, clubID)
	if err != nil {
		return nil, nil, nil, err
	}

	s.log.Info("game participation accepted",
		"game_id", gameID,
		"club_id", clubID,
		"player_id", player.ID,
		"tg_user_id", tgUserID,
	)

	return player, game, notifyIDs, nil
}

// DeclineGameParticipation allows a player to decline a game invitation.
// The player must have an invited or accepted status for the game.
func (s *Service) DeclineGameParticipation(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, []int64, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return nil, nil, errors.New("отказаться можно только для игр в статусе planned")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return nil, nil, errors.New("вы не приглашены в эту игру")
	}

	if participant.Status == "declined" {
		return nil, nil, errors.New("вы уже отказались от участия")
	}

	if err := s.repos.GameParticipants.UpdateStatus(ctx, gameID, player.ID, "declined"); err != nil {
		return nil, nil, fmt.Errorf("failed to decline game participation: %w", err)
	}

	// Get owner and admin tg_user_ids for notification.
	notifyIDs, err := s.getOwnerAndAdminTgUserIDs(ctx, clubID)
	if err != nil {
		return nil, nil, err
	}

	s.log.Info("game participation declined",
		"game_id", gameID,
		"club_id", clubID,
		"player_id", player.ID,
		"tg_user_id", tgUserID,
	)

	return game, notifyIDs, nil
}

// ConfirmGameParticipation allows owner/admin to confirm a player's accepted
// invitation to a game. The player must have accepted status.
func (s *Service) ConfirmGameParticipation(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64) (*domain.Player, *domain.Game, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermManageGameParticipants); err != nil {
		return nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return nil, nil, errors.New("подтверждение доступно только для игр в статусе planned")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID)
	if err != nil {
		return nil, nil, errors.New("игрок не является участником игры")
	}

	if participant.Status != "accepted" {
		return nil, nil, errors.New("игрок не ожидает подтверждения")
	}

	if err := s.repos.GameParticipants.UpdateStatus(ctx, gameID, playerID, "confirmed"); err != nil {
		return nil, nil, fmt.Errorf("failed to confirm game participation: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, nil, err
	}

	s.log.Info("game participation confirmed",
		"game_id", gameID,
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, game, nil
}

// RemoveGameParticipant removes a player from a game. Only owner/admin can do this.
// The game must be in planned status.
func (s *Service) RemoveGameParticipant(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64) (*domain.Player, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermManageGameParticipants); err != nil {
		return nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "planned" {
		return nil, errors.New("изменение состава доступно только для игр в статусе planned")
	}

	_, err = s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID)
	if err != nil {
		return nil, errors.New("игрок не является участником игры")
	}

	if err := s.repos.GameParticipants.Delete(ctx, gameID, playerID); err != nil {
		return nil, fmt.Errorf("failed to remove game participant: %w", err)
	}

	player, err := s.repos.Players.GetByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	s.log.Info("game participant removed",
		"game_id", gameID,
		"club_id", clubID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return player, nil
}

// GetGameParticipants returns all participants of a game with player info.
func (s *Service) GetGameParticipants(ctx context.Context, tgUserID int64, clubID int64, gameID int64) ([]*domain.GameParticipantWithPlayer, error) {
	if err := s.checkPermission(ctx, tgUserID, clubID, PermViewClub); err != nil {
		return nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	return s.repos.GameParticipants.GetByGameWithPlayers(ctx, gameID)
}

// GetGameParticipant returns a specific game participant with player info.
func (s *Service) GetGameParticipant(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.GameParticipantWithPlayer, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return nil, errors.New("вы не являетесь участником этой игры")
	}

	p, err := s.repos.Players.GetByID(ctx, participant.PlayerID)
	if err != nil {
		return nil, err
	}

	return &domain.GameParticipantWithPlayer{
		GameParticipant: *participant,
		Player:          *p,
	}, nil
}

// --- Phase 04: Game start process ---

// checkGameAccess verifies that the Telegram user is the banker, owner, or admin
// of the club that owns the game. It returns the game and the club member.
func (s *Service) checkGameAccess(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, *domain.ClubMember, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, fmt.Errorf("access denied: user not found: %w", err)
	}

	member, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, player.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("access denied: user is not a member of this club: %w", err)
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	// Owner or admin has access.
	if member.Role == "owner" || member.Role == "admin" {
		return game, member, nil
	}

	// Banker has access (banker_id references club_members.id).
	if member.ID == game.BankerID {
		return game, member, nil
	}

	s.log.Warn("access denied",
		"tg_user_id", tgUserID,
		"club_id", clubID,
		"game_id", gameID,
		"role", member.Role,
	)
	return nil, nil, errors.New("access denied: insufficient permissions")
}

// StartGame transitions a game from planned to active.
// The requesting user must be the banker, owner, or admin.
// On start: status → active, start_time → now, Buy-in registered for all
// confirmed/accepted participants, timer started if duration is set.
func (s *Service) StartGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, error) {
	game, member, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != "planned" {
		return nil, errors.New("игра должна быть в статусе planned")
	}

	// Check constraint 7.15: no active games in the club.
	_, err = s.repos.Games.GetActiveByClub(ctx, clubID)
	if err == nil {
		return nil, errors.New("в клубе уже есть активная игра")
	}

	// Transition to active and record actual start time.
	startTime := time.Now()
	if err := s.repos.Games.StartGame(ctx, gameID, startTime); err != nil {
		return nil, fmt.Errorf("failed to start game: %w", err)
	}

	// Register Buy-in for all confirmed/accepted participants.
	participants, err := s.repos.GameParticipants.GetByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game participants: %w", err)
	}

	for _, p := range participants {
		if p.Status == "confirmed" || p.Status == "accepted" {
			if err := s.repos.GameParticipants.RegisterBuyIn(ctx, gameID, p.PlayerID, 1); err != nil {
				return nil, fmt.Errorf("failed to register buy-in for player %d: %w", p.PlayerID, err)
			}
			// Record buy-in event.
			buyInAmount := game.BuyInAmount
			event := &domain.Event{
				GameID:    gameID,
				PlayerID:  p.PlayerID,
				Type:      "buy_in",
				NewValue:  &buyInAmount,
				CreatedBy: member.ID,
			}
			if _, err := s.repos.Events.Create(ctx, event); err != nil {
				s.log.Warn("failed to create buy-in event", "error", err)
			}
		}
	}

	s.log.Info("game started",
		"game_id", gameID,
		"club_id", clubID,
		"tg_user_id", tgUserID,
		"start_time", startTime,
	)

	return s.repos.Games.GetByID(ctx, gameID)
}

// RegisterRebuy registers a rebuy for a player in an active game.
// Only the banker, owner, or admin can perform this action.
// Rebuy is only allowed if rebuy_allowed is true and max_rebuys is not exceeded.
func (s *Service) RegisterRebuy(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64) error {
	game, member, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("регистрация rebuy доступна только для активных игр")
	}

	if !game.RebuyAllowed {
		return errors.New("rebuy не разрешен для этой игры")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID)
	if err != nil {
		return errors.New("игрок не является участником игры")
	}

	// Check max_rebuys constraint.
	if game.MaxRebuys != nil && participant.RebuyCount >= *game.MaxRebuys {
		return errors.New("максимальное количество rebuy достигнуто")
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	newRebuyCount := participant.RebuyCount + 1
	if err := s.repos.GameParticipants.RegisterRebuy(ctx, gameID, playerID, newRebuyCount); err != nil {
		return fmt.Errorf("failed to register rebuy: %w", err)
	}

	// Record rebuy event.
	event := &domain.Event{
		GameID:    gameID,
		PlayerID:  playerID,
		Type:      "rebuy",
		NewValue:  &rebuyPrice,
		CreatedBy: member.ID,
	}
	if _, err := s.repos.Events.Create(ctx, event); err != nil {
		s.log.Warn("failed to create rebuy event", "error", err)
	}

	s.log.Info("rebuy registered",
		"game_id", gameID,
		"player_id", playerID,
		"rebuy_count", newRebuyCount,
		"tg_user_id", tgUserID,
	)

	return nil
}

// FixRebuy corrects an erroneous rebuy for a player.
// The banker, owner, or admin can adjust the rebuy_count.
// The rebuy amount is recalculated from the count and game.rebuy_price.
// A correction event is recorded to preserve history.
func (s *Service) FixRebuy(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64, newRebuyCount int) error {
	game, member, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("исправление rebuy доступно только для активных игр")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID)
	if err != nil {
		return errors.New("игрок не является участником игры")
	}

	oldRebuyCount := participant.RebuyCount

	participant.RebuyCount = newRebuyCount

	if err := s.repos.GameParticipants.Update(ctx, participant); err != nil {
		return fmt.Errorf("failed to fix rebuy: %w", err)
	}

	// Calculate old and new rebuy amounts from counts.
	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}
	oldRebuyAmount := float64(oldRebuyCount) * rebuyPrice
	newRebuyAmount := float64(newRebuyCount) * rebuyPrice

	// Record correction event.
	metadata := map[string]interface{}{
		"old_rebuy_count": oldRebuyCount,
		"new_rebuy_count": newRebuyCount,
		"old_rebuy_amount": oldRebuyAmount,
		"new_rebuy_amount": newRebuyAmount,
	}
	event := &domain.Event{
		GameID:    gameID,
		PlayerID:  playerID,
		Type:      "correction",
		OldValue:  &oldRebuyAmount,
		NewValue:  &newRebuyAmount,
		Metadata:  metadata,
		CreatedBy: member.ID,
	}
	if _, err := s.repos.Events.Create(ctx, event); err != nil {
		s.log.Warn("failed to create correction event", "error", err)
	}

	s.log.Info("rebuy fixed",
		"game_id", gameID,
		"player_id", playerID,
		"old_rebuy_count", oldRebuyCount,
		"new_rebuy_count", newRebuyCount,
		"tg_user_id", tgUserID,
	)

	return nil
}

// UpdateCurrentStack allows a player to update their own current chip stack
// during an active game. The value must not be negative.
// The current stack is stored as a chips_set event in the events table.
func (s *Service) UpdateCurrentStack(ctx context.Context, tgUserID int64, clubID int64, gameID int64, stack float64) error {
	if stack < 0 {
		return errors.New("количество фишек не может быть отрицательным")
	}

	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.ClubID != clubID {
		return errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "active" {
		return errors.New("ввод стека доступен только для активных игр")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return errors.New("вы не являетесь участником этой игры")
	}

	// Record chips_set event.
	event := &domain.Event{
		GameID:    gameID,
		PlayerID:  player.ID,
		Type:      "chips_set",
		NewValue:  &stack,
		CreatedBy: participant.ID,
	}
	if _, err := s.repos.Events.Create(ctx, event); err != nil {
		return fmt.Errorf("failed to record chips_set event: %w", err)
	}

	s.log.Info("current stack updated",
		"game_id", gameID,
		"player_id", player.ID,
		"stack", stack,
	)

	return nil
}

// AddPlayerToGame adds a new player to an active Cash game.
// The requesting user must be the banker, owner, or admin.
// The player must be an active club member, not already in the game,
// and max_players must not be exceeded.
// A Buy-in is automatically registered for the new player.
func (s *Service) AddPlayerToGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64) error {
	game, member, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("добавление игрока доступно только для активных игр")
	}

	if game.GameType != "cash" {
		return errors.New("добавление игрока доступно только для Cash-игр")
	}

	// Check player is an active club member.
	playerMember, err := s.repos.ClubMembers.GetByClubAndPlayer(ctx, clubID, playerID)
	if err != nil {
		return errors.New("игрок не является участником клуба")
	}
	if playerMember.Status != "active" {
		return errors.New("игрок не является активным участником клуба")
	}

	// Check player is not already in the game.
	existing, _ := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID)
	if existing != nil {
		return errors.New("игрок уже является участником этой игры")
	}

	// Check max_players constraint.
	participants, err := s.repos.GameParticipants.GetByGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game participants: %w", err)
	}
	if len(participants) >= game.MaxPlayers {
		return errors.New("достигнуто максимальное количество игроков")
	}

	// Create participant with initial Buy-in.
	participant := &domain.GameParticipant{
		GameID:     gameID,
		PlayerID:   playerID,
		BuyInCount: 1,
		Status:     "confirmed",
	}
	if _, err := s.repos.GameParticipants.Create(ctx, participant); err != nil {
		return fmt.Errorf("failed to create game participant: %w", err)
	}

	// Record buy-in event.
	buyInAmount := game.BuyInAmount
	event := &domain.Event{
		GameID:    gameID,
		PlayerID:  playerID,
		Type:      "buy_in",
		NewValue:  &buyInAmount,
		CreatedBy: member.ID,
	}
	if _, err := s.repos.Events.Create(ctx, event); err != nil {
		s.log.Warn("failed to create buy-in event for added player", "error", err)
	}

	s.log.Info("player added to game",
		"game_id", gameID,
		"player_id", playerID,
		"tg_user_id", tgUserID,
	)

	return nil
}

// PauseTimer pauses the game timer. Only available for Cash games with a duration.
// The requesting user must be the banker, owner, or admin.
func (s *Service) PauseTimer(ctx context.Context, tgUserID int64, clubID int64, gameID int64) error {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("управление таймером доступно только для активных игр")
	}

	if game.Duration == nil {
		return errors.New("у этой игры нет таймера")
	}

	if game.TimerPausedAt != nil {
		return errors.New("таймер уже приостановлен")
	}

	now := time.Now()
	if err := s.repos.Games.UpdateTimer(ctx, gameID, &now, game.TimerPausedDuration, false); err != nil {
		return fmt.Errorf("failed to pause timer: %w", err)
	}

	s.log.Info("timer paused",
		"game_id", gameID,
		"tg_user_id", tgUserID,
	)

	return nil
}

// ResumeTimer resumes a paused game timer.
// The requesting user must be the banker, owner, or admin.
func (s *Service) ResumeTimer(ctx context.Context, tgUserID int64, clubID int64, gameID int64) error {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("управление таймером доступно только для активных игр")
	}

	if game.Duration == nil {
		return errors.New("у этой игры нет таймера")
	}

	if game.TimerPausedAt == nil {
		return errors.New("таймер не приостановлен")
	}

	// Calculate additional paused duration.
	pausedDuration := time.Duration(0)
	if game.TimerPausedDuration != nil {
		pausedDuration = *game.TimerPausedDuration
	}
	pausedDuration += time.Since(*game.TimerPausedAt)

	if err := s.repos.Games.UpdateTimer(ctx, gameID, nil, &pausedDuration, false); err != nil {
		return fmt.Errorf("failed to resume timer: %w", err)
	}

	s.log.Info("timer resumed",
		"game_id", gameID,
		"tg_user_id", tgUserID,
	)

	return nil
}

// ExtendGame extends the game duration by the specified amount.
// The requesting user must be the banker, owner, or admin.
func (s *Service) ExtendGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64, additionalDuration time.Duration) error {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("продление доступно только для активных игр")
	}

	if game.Duration == nil {
		return errors.New("у этой игры нет таймера")
	}

	if err := s.repos.Games.ExtendDuration(ctx, gameID, additionalDuration); err != nil {
		return fmt.Errorf("failed to extend game duration: %w", err)
	}

	s.log.Info("game extended",
		"game_id", gameID,
		"additional_duration", additionalDuration,
		"tg_user_id", tgUserID,
	)

	return nil
}

// GetGameMonitor returns the game and all participants for the banker/owner/admin
// game monitor view.
func (s *Service) GetGameMonitor(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, []*domain.GameParticipantWithPlayer, error) {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.Status != "active" {
		return nil, nil, errors.New("монитор доступен только для активных игр")
	}

	participants, err := s.repos.GameParticipants.GetByGameWithPlayers(ctx, gameID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game participants: %w", err)
	}

	return game, participants, nil
}

// GetPlayerGameMonitor returns the game and the requesting player's participant data
// for the player's game monitor view.
func (s *Service) GetPlayerGameMonitor(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, *domain.GameParticipantWithPlayer, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "active" {
		return nil, nil, errors.New("монитор доступен только для активных игр")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return nil, nil, errors.New("вы не являетесь участником этой игры")
	}

	p, err := s.repos.Players.GetByID(ctx, participant.PlayerID)
	if err != nil {
		return nil, nil, err
	}

	return game, &domain.GameParticipantWithPlayer{
		GameParticipant: *participant,
		Player:          *p,
	}, nil
}

// GetPlayerGameStats returns the game and the requesting player's participant data
// for the "Моя статистика" view.
func (s *Service) GetPlayerGameStats(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, *domain.GameParticipant, error) {
	player, err := s.repos.Players.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		return nil, nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.ClubID != clubID {
		return nil, nil, errors.New("игра не принадлежит этому клубу")
	}

	if game.Status != "active" {
		return nil, nil, errors.New("статистика доступна только для активных игр")
	}

	participant, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, player.ID)
	if err != nil {
		return nil, nil, errors.New("вы не являетесь участником этой игры")
	}

	return game, participant, nil
}

// GetGameStatistics returns the game and all participants for the game statistics view.
// The requesting user must be the banker, owner, or admin.
func (s *Service) GetGameStatistics(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*domain.Game, []*domain.GameParticipantWithPlayer, error) {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.Status != "active" {
		return nil, nil, errors.New("статистика доступна только для активных игр")
	}

	participants, err := s.repos.GameParticipants.GetByGameWithPlayers(ctx, gameID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game participants: %w", err)
	}

	return game, participants, nil
}

// GetRebuyEvents returns all rebuy events for a player in a game.
// The requesting user must be the banker, owner, or admin.
func (s *Service) GetRebuyEvents(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64) ([]*domain.Event, error) {
	_, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return nil, err
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != "active" {
		return nil, errors.New("управление rebuy доступно только для активных игр")
	}

	events, err := s.repos.Events.GetByGameAndPlayer(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Filter to only rebuy events.
	var rebuyEvents []*domain.Event
	for _, e := range events {
		if e.Type == "rebuy" {
			rebuyEvents = append(rebuyEvents, e)
		}
	}

	return rebuyEvents, nil
}

// GetActiveGameByPlayer returns the active game for a club if the player
// is a participant in it. Used to check constraint 7.6.
func (s *Service) GetActiveGameByPlayer(ctx context.Context, clubID int64, playerID int64) (*domain.Game, error) {
	game, err := s.repos.Games.GetActiveByClub(ctx, clubID)
	if err != nil {
		return nil, err
	}

	if _, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, game.ID, playerID); err != nil {
		return nil, errors.New("игрок не участвует в активной игре")
	}

	return game, nil
}

// GetCurrentStacks returns a map of playerID to current stack for all participants
// in a game, based on the last chips_set event for each player.
func (s *Service) GetCurrentStacks(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (map[int64]float64, error) {
	if _, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID); err != nil {
		// Fall back to permission check for non-banker users.
		if err := s.checkPermission(ctx, tgUserID, clubID, PermViewClub); err != nil {
			return nil, err
		}
	}

	game, err := s.repos.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.ClubID != clubID {
		return nil, errors.New("игра не принадлежит этому клубу")
	}

	return s.repos.Events.GetLastChipsSetByGame(ctx, gameID)
}

// TimerNotification contains a game with expired timer and the Telegram user IDs
// of the banker and owner/admin who should be notified.
type TimerNotification struct {
	Game       *domain.Game
	TgUserIDs  []int64
}

// CheckExpiredTimers returns games with expired timers and the Telegram user IDs
// of the banker and owner/admin who should be notified. It also marks the
// notification as sent (timer_notified = true).
func (s *Service) CheckExpiredTimers(ctx context.Context) ([]*TimerNotification, error) {
	games, err := s.repos.Games.GetExpiredTimerGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired timer games: %w", err)
	}

	var notifications []*TimerNotification
	for _, game := range games {
		// Get all club members with player info.
		members, err := s.repos.ClubMembers.GetByClubWithPlayers(ctx, game.ClubID)
		if err != nil {
			s.log.Warn("failed to get club members", "error", err, "game_id", game.ID, "club_id", game.ClubID)
			continue
		}

		var tgUserIDs []int64
		seen := map[int64]bool{}

		for _, m := range members {
			// Banker gets notified.
			if m.ID == game.BankerID && m.Player.TgUserID != nil {
				if !seen[*m.Player.TgUserID] {
					tgUserIDs = append(tgUserIDs, *m.Player.TgUserID)
					seen[*m.Player.TgUserID] = true
				}
			}
			// Owner/Admin get notified.
			if (m.Role == "owner" || m.Role == "admin") && m.Player.TgUserID != nil {
				if !seen[*m.Player.TgUserID] {
					tgUserIDs = append(tgUserIDs, *m.Player.TgUserID)
					seen[*m.Player.TgUserID] = true
				}
			}
		}

		// Mark notification as sent.
		if err := s.repos.Games.UpdateTimer(ctx, game.ID, game.TimerPausedAt, game.TimerPausedDuration, true); err != nil {
			s.log.Warn("failed to mark timer notified", "error", err, "game_id", game.ID)
		}

		notifications = append(notifications, &TimerNotification{
			Game:      game,
			TgUserIDs: tgUserIDs,
		})
	}

	return notifications, nil
}

// --- Phase 05: Game end ---

// GameBankCheck holds the result of a bank correctness check for a Cash game.
type GameBankCheck struct {
	TotalBank   float64
	TotalPayout float64
	Difference  float64
	Mismatch    bool
}

// participantResult holds the calculated results for a single participant
// during game finish.
type participantResult struct {
	participant   *domain.GameParticipant
	buyInAmount   float64
	rebuyAmount   float64
	totalInvested float64
	payoutAmount  float64
	profit        float64
	roi           float64
}

// SetChipsEnd allows the banker/owner/admin to set the final chip stack
// (chips_end) for a participant in an active game. The value must not be negative.
func (s *Service) SetChipsEnd(ctx context.Context, tgUserID int64, clubID int64, gameID int64, playerID int64, chipsEnd float64) error {
	if chipsEnd < 0 {
		return errors.New("количество фишек не может быть отрицательным")
	}

	game, member, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("ввод chips_end доступен только для активных игр")
	}

	// Verify the player is a participant.
	if _, err := s.repos.GameParticipants.GetByGameAndPlayer(ctx, gameID, playerID); err != nil {
		return errors.New("игрок не является участником игры")
	}

	if err := s.repos.GameParticipants.UpdateChipsEnd(ctx, gameID, playerID, chipsEnd); err != nil {
		return fmt.Errorf("failed to update chips_end: %w", err)
	}

	// Record a chips_set event for the final stack.
	event := &domain.Event{
		GameID:    gameID,
		PlayerID:  playerID,
		Type:      "chips_set",
		NewValue:  &chipsEnd,
		CreatedBy: member.ID,
	}
	if _, err := s.repos.Events.Create(ctx, event); err != nil {
		s.log.Warn("failed to create chips_set event", "error", err)
	}

	s.log.Info("chips_end set",
		"game_id", gameID,
		"player_id", playerID,
		"chips_end", chipsEnd,
		"tg_user_id", tgUserID,
	)

	return nil
}

// CheckGameBank calculates the total bank (sum of all buy-in and rebuy amounts)
// and the total payout (sum of chips_end × chip_value) for a game. It returns
// whether they match and the difference.
func (s *Service) CheckGameBank(ctx context.Context, tgUserID int64, clubID int64, gameID int64) (*GameBankCheck, error) {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != "active" {
		return nil, errors.New("проверка банка доступна только для активных игр")
	}

	participants, err := s.repos.GameParticipants.GetByGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game participants: %w", err)
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	totalBank := 0.0
	totalPayout := 0.0

	for _, p := range participants {
		buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
		rebuyAmount := float64(p.RebuyCount) * rebuyPrice
		totalBank += buyInAmount + rebuyAmount

		if p.ChipsEnd != nil {
			totalPayout += *p.ChipsEnd * game.ChipValue
		}
	}

	diff := totalPayout - totalBank
	return &GameBankCheck{
		TotalBank:   totalBank,
		TotalPayout: totalPayout,
		Difference:  diff,
		Mismatch:    diff != 0,
	}, nil
}

// FinishGame completes a Cash game: calculates payout_amount, profit, ROI,
// place, and winner for each participant, distributes any bank mismatch among
// players with positive profit, updates player statistics, records the end
// time, and transitions the game to finished status.
//
// For Tournament games, this method returns an error as Tournament result
// logic is not implemented in Phase 5.
func (s *Service) FinishGame(ctx context.Context, tgUserID int64, clubID int64, gameID int64) error {
	game, _, err := s.checkGameAccess(ctx, tgUserID, clubID, gameID)
	if err != nil {
		return err
	}

	if game.Status != "active" {
		return errors.New("завершение игры доступно только для активных игр")
	}

	// Tournament result logic is not implemented in Phase 5.
	if game.GameType != "cash" {
		return errors.New("завершение tournament игр не реализовано")
	}

	participants, err := s.repos.GameParticipants.GetByGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game participants: %w", err)
	}

	// Constraint 7.9: all players' results must be entered.
	for _, p := range participants {
		if p.ChipsEnd == nil {
			return errors.New("не введены результаты всех игроков")
		}
	}

	rebuyPrice := 0.0
	if game.RebuyPrice != nil {
		rebuyPrice = *game.RebuyPrice
	}

	results := make([]*participantResult, 0, len(participants))
	totalBank := 0.0
	totalPayout := 0.0

	for _, p := range participants {
		buyInAmount := float64(p.BuyInCount) * game.BuyInAmount
		rebuyAmount := float64(p.RebuyCount) * rebuyPrice
		totalInvested := buyInAmount + rebuyAmount
		payoutAmount := *p.ChipsEnd * game.ChipValue
		profit := payoutAmount - totalInvested

		var roi float64
		if totalInvested > 0 {
			roi = (profit / totalInvested) * 100
		}

		totalBank += totalInvested
		totalPayout += payoutAmount

		results = append(results, &participantResult{
			participant:   p,
			buyInAmount:   buyInAmount,
			rebuyAmount:   rebuyAmount,
			totalInvested: totalInvested,
			payoutAmount:  payoutAmount,
			profit:        profit,
			roi:           roi,
		})
	}

	// Constraint 7.10 / Exception 8.7: if the bank doesn't match, distribute
	// the difference equally among players with positive profit.
	diff := totalPayout - totalBank
	if diff != 0 {
		var positiveProfit []*participantResult
		for _, r := range results {
			if r.profit > 0 {
				positiveProfit = append(positiveProfit, r)
			}
		}

		if len(positiveProfit) > 0 {
			adjustment := -diff / float64(len(positiveProfit))
			for _, r := range positiveProfit {
				r.payoutAmount += adjustment
				r.profit = r.payoutAmount - r.totalInvested
				if r.totalInvested > 0 {
					r.roi = (r.profit / r.totalInvested) * 100
				}
			}
		}

		s.log.Warn("bank mismatch resolved",
			"game_id", gameID,
			"difference", diff,
			"adjusted_players", len(positiveProfit),
		)
	}

	// Sort by descending profit, then by descending ROI.
	sort.Slice(results, func(i, j int) bool {
		if results[i].profit != results[j].profit {
			return results[i].profit > results[j].profit
		}
		return results[i].roi > results[j].roi
	})

	// Assign places. Players with equal profit and ROI get the same place.
	place := 0
	for i, r := range results {
		if i == 0 || r.profit != results[i-1].profit || r.roi != results[i-1].roi {
			place = i + 1
		}
		r.participant.PayoutAmount = &r.payoutAmount
		r.participant.Place = &place
	}

	// Save calculated results to game_participants.
	for _, r := range results {
		if err := s.repos.GameParticipants.Update(ctx, r.participant); err != nil {
			return fmt.Errorf("failed to save game participant results: %w", err)
		}
	}

	// Update player statistics for each participant.
	for _, r := range results {
		if err := s.updatePlayerStatistics(ctx, r.participant.PlayerID, game.ClubID, r); err != nil {
			s.log.Warn("failed to update player statistics",
				"error", err,
				"player_id", r.participant.PlayerID,
				"game_id", gameID,
			)
		}
	}

	// Transition game to finished and record end time.
	endTime := time.Now()
	if err := s.repos.Games.FinishGame(ctx, gameID, endTime); err != nil {
		return fmt.Errorf("failed to finish game: %w", err)
	}

	s.log.Info("game finished",
		"game_id", gameID,
		"club_id", clubID,
		"tg_user_id", tgUserID,
		"end_time", endTime,
	)

	return nil
}

// updatePlayerStatistics updates the cached aggregate statistics for a player
// after a game finishes. It reads the existing statistics (if any), adds the
// current game's values, recalculates derived metrics, and upserts.
func (s *Service) updatePlayerStatistics(ctx context.Context, playerID, clubID int64, r *participantResult) error {
	existing, err := s.repos.PlayerStatistics.GetByPlayerAndClub(ctx, playerID, clubID)
	if err != nil {
		// No existing statistics — start fresh.
		existing = &domain.PlayerStatistics{
			PlayerID: playerID,
			ClubID:   clubID,
		}
	}

	place := 0
	if r.participant.Place != nil {
		place = *r.participant.Place
	}

	stats := &domain.PlayerStatistics{
		PlayerID:         playerID,
		ClubID:           clubID,
		TotalGames:       existing.TotalGames + 1,
		TotalBuyInAmount: existing.TotalBuyInAmount + r.buyInAmount,
		TotalRebuyAmount: existing.TotalRebuyAmount + r.rebuyAmount,
		TotalRebuysCount: existing.TotalRebuysCount + r.participant.RebuyCount,
		TotalInvested:    existing.TotalInvested + r.totalInvested,
		TotalChips:       existing.TotalChips + *r.participant.ChipsEnd,
		TotalProfit:      existing.TotalProfit + r.profit,
		BiggestWin:       existing.BiggestWin,
		BiggestLoss:      existing.BiggestLoss,
		GamesWon:         existing.GamesWon,
		Podiums:          existing.Podiums,
		ROI:              existing.ROI,
		ITM:              existing.ITM,
	}

	if r.profit > existing.BiggestWin {
		stats.BiggestWin = r.profit
	}
	if r.profit < existing.BiggestLoss {
		stats.BiggestLoss = r.profit
	}
	if place == 1 {
		stats.GamesWon = existing.GamesWon + 1
	}

	// Recalculate ROI.
	if stats.TotalInvested > 0 {
		stats.ROI = (stats.TotalProfit / stats.TotalInvested) * 100
	}

	return s.repos.PlayerStatistics.Upsert(ctx, stats)
}

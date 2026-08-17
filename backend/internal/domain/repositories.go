package domain

import "context"

// ClubRepository defines operations for club persistence.
type ClubRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, club *Club) (int64, error)
	GetByID(ctx context.Context, id int64) (*Club, error)
	GetByOwner(ctx context.Context, playerID int64) ([]*Club, error)
	GetByPlayer(ctx context.Context, playerID int64) ([]*Club, error)
	GetByTgChatID(ctx context.Context, tgChatID int64) (*Club, error)
	UpdateName(ctx context.Context, id int64, name string) error
	BindTgChatID(ctx context.Context, clubID, tgChatID int64) error
	Delete(ctx context.Context, id int64) error
}

// PlayerRepository defines operations for player persistence.
type PlayerRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, player *Player) (int64, error)
	GetByID(ctx context.Context, id int64) (*Player, error)
	GetByTgUserID(ctx context.Context, tgUserID int64) (*Player, error)
	GetByNickname(ctx context.Context, nickname string) (*Player, error)
	UpdateLastSeen(ctx context.Context, id int64) error
}

// ClubMemberRepository defines operations for club member persistence.
type ClubMemberRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, member *ClubMember) (int64, error)
	GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*ClubMember, error)
	GetByClubWithPlayers(ctx context.Context, clubID int64) ([]*ClubMemberWithPlayer, error)
	UpdateRole(ctx context.Context, clubID, playerID int64, role string) error
	UpdateStatus(ctx context.Context, clubID, playerID int64, status string) error
	UpdateAccepted(ctx context.Context, clubID, playerID int64, accepted bool) error
}

// GameRepository defines operations for game persistence.
type GameRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, game *Game) (int64, error)
	GetByID(ctx context.Context, id int64) (*Game, error)
	GetByClub(ctx context.Context, clubID int64) ([]*Game, error)
	GetActiveByClub(ctx context.Context, clubID int64) (*Game, error)
	Update(ctx context.Context, game *Game) error
	Cancel(ctx context.Context, id int64) error
}

// GameParticipantRepository defines operations for game participant persistence.
type GameParticipantRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, participant *GameParticipant) (int64, error)
	GetByID(ctx context.Context, id int64) (*GameParticipant, error)
	GetByGame(ctx context.Context, gameID int64) ([]*GameParticipant, error)
	GetByGameWithPlayers(ctx context.Context, gameID int64) ([]*GameParticipantWithPlayer, error)
	GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) (*GameParticipant, error)
	Update(ctx context.Context, participant *GameParticipant) error
	UpdateStatus(ctx context.Context, gameID, playerID int64, status string) error
	Delete(ctx context.Context, gameID, playerID int64) error
}

// EventRepository defines operations for event log persistence.
type EventRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, event *Event) (int64, error)
}

// PlayerStatisticsRepository defines operations for player statistics persistence.
type PlayerStatisticsRepository interface {
	Ping(ctx context.Context) error
}

// Repositories groups all repository interfaces.
type Repositories struct {
	Clubs            ClubRepository
	Players          PlayerRepository
	ClubMembers      ClubMemberRepository
	Games            GameRepository
	GameParticipants GameParticipantRepository
	Events           EventRepository
	PlayerStatistics PlayerStatisticsRepository
}

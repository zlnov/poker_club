package domain

import "context"

// ClubRepository defines operations for club persistence.
type ClubRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, club *Club) (int64, error)
	GetByID(ctx context.Context, id int64) (*Club, error)
	GetByOwner(ctx context.Context, playerID int64) ([]*Club, error)
	GetByTgChatID(ctx context.Context, tgChatID int64) (*Club, error)
	UpdateName(ctx context.Context, id int64, name string) error
	BindTgChatID(ctx context.Context, clubID, tgChatID int64) error
	Delete(ctx context.Context, id int64) error
}

// PlayerRepository defines operations for player persistence.
type PlayerRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, player *Player) (int64, error)
	GetByTgUserID(ctx context.Context, tgUserID int64) (*Player, error)
	UpdateLastSeen(ctx context.Context, id int64) error
}

// ClubMemberRepository defines operations for club member persistence.
type ClubMemberRepository interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, member *ClubMember) (int64, error)
	GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*ClubMember, error)
}

// GameRepository defines operations for game persistence.
type GameRepository interface {
	Ping(ctx context.Context) error
}

// GameParticipantRepository defines operations for game participant persistence.
type GameParticipantRepository interface {
	Ping(ctx context.Context) error
}

// EventRepository defines operations for event log persistence.
type EventRepository interface {
	Ping(ctx context.Context) error
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

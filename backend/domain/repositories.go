package domain

import (
	"context"
)

// ClubRepository interface for club persistence
type ClubRepository interface {
	Create(ctx context.Context, club *Club) error
	GetByID(ctx context.Context, id int64) (*Club, error)
	GetAll(ctx context.Context) ([]*Club, error)
	Update(ctx context.Context, club *Club) error
	Delete(ctx context.Context, id int64) error
}

// PlayerRepository interface for player persistence
type PlayerRepository interface {
	Create(ctx context.Context, player *Player) error
	GetByID(ctx context.Context, id int64) (*Player, error)
	GetByPhone(ctx context.Context, phoneNumber string) (*Player, error)
	GetAll(ctx context.Context) ([]*Player, error)
	Update(ctx context.Context, player *Player) error
	Delete(ctx context.Context, id int64) error
}

// ClubMemberRepository interface for club member persistence
type ClubMemberRepository interface {
	Create(ctx context.Context, member *ClubMember) error
	GetByID(ctx context.Context, id int64) (*ClubMember, error)
	GetByClubID(ctx context.Context, clubID int64) ([]*ClubMember, error)
	GetByPlayerID(ctx context.Context, playerID int64) ([]*ClubMember, error)
	GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*ClubMember, error)
	Update(ctx context.Context, member *ClubMember) error
	Delete(ctx context.Context, id int64) error
}

// GameRepository interface for game persistence
type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id int64) (*Game, error)
	GetByClubID(ctx context.Context, clubID int64) ([]*Game, error)
	GetByBankerID(ctx context.Context, bankerID int64) ([]*Game, error)
	GetByClubIDWithFilters(ctx context.Context, clubID int64, status *string, limit, offset int) ([]*Game, error)
	CountByClubID(ctx context.Context, clubID int64, status *string) (int64, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id int64) error
}

// GameParticipantRepository interface for game participant persistence
type GameParticipantRepository interface {
	Create(ctx context.Context, participant *GameParticipant) error
	GetByID(ctx context.Context, id int64) (*GameParticipant, error)
	GetByGameID(ctx context.Context, gameID int64) ([]*GameParticipant, error)
	GetByPlayerID(ctx context.Context, playerID int64) ([]*GameParticipant, error)
	GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) (*GameParticipant, error)
	Update(ctx context.Context, participant *GameParticipant) error
	Delete(ctx context.Context, id int64) error
	LockForUpdate(ctx context.Context, gameID int64) ([]*GameParticipant, error)
}

// EventRepository interface for event persistence
type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id int64) (*Event, error)
	GetByGameID(ctx context.Context, gameID int64) ([]*Event, error)
	GetByPlayerID(ctx context.Context, playerID int64) ([]*Event, error)
	GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) ([]*Event, error)
}

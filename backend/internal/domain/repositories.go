package domain

import "context"

// ClubRepository defines operations for club persistence.
type ClubRepository interface {
	Ping(ctx context.Context) error
}

// PlayerRepository defines operations for player persistence.
type PlayerRepository interface {
	Ping(ctx context.Context) error
}

// ClubMemberRepository defines operations for club member persistence.
type ClubMemberRepository interface {
	Ping(ctx context.Context) error
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
	Clubs             ClubRepository
	Players           PlayerRepository
	ClubMembers       ClubMemberRepository
	Games             GameRepository
	GameParticipants  GameParticipantRepository
	Events            EventRepository
	PlayerStatistics  PlayerStatisticsRepository
}

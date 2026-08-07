package postgres

import (
	"context"

	"poker-club/backend/internal/domain"
)

// clubRepository implements domain.ClubRepository.
type clubRepository struct {
	db *DB
}

func (r *clubRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// playerRepository implements domain.PlayerRepository.
type playerRepository struct {
	db *DB
}

func (r *playerRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// clubMemberRepository implements domain.ClubMemberRepository.
type clubMemberRepository struct {
	db *DB
}

func (r *clubMemberRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// gameRepository implements domain.GameRepository.
type gameRepository struct {
	db *DB
}

func (r *gameRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// gameParticipantRepository implements domain.GameParticipantRepository.
type gameParticipantRepository struct {
	db *DB
}

func (r *gameParticipantRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// eventRepository implements domain.EventRepository.
type eventRepository struct {
	db *DB
}

func (r *eventRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// playerStatisticsRepository implements domain.PlayerStatisticsRepository.
type playerStatisticsRepository struct {
	db *DB
}

func (r *playerStatisticsRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// NewRepositories creates all repository implementations and returns them grouped.
func NewRepositories(db *DB) *domain.Repositories {
	return &domain.Repositories{
		Clubs:            &clubRepository{db: db},
		Players:          &playerRepository{db: db},
		ClubMembers:      &clubMemberRepository{db: db},
		Games:            &gameRepository{db: db},
		GameParticipants: &gameParticipantRepository{db: db},
		Events:           &eventRepository{db: db},
		PlayerStatistics: &playerStatisticsRepository{db: db},
	}
}

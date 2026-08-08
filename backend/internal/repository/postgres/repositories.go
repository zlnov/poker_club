package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"poker-club/backend/internal/domain"
)

// clubRepository implements domain.ClubRepository.
type clubRepository struct {
	db *DB
}

func (r *clubRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *clubRepository) Create(ctx context.Context, club *domain.Club) (int64, error) {
	query := `INSERT INTO clubs (tg_chat_id, name) VALUES ($1, $2) RETURNING id`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query, club.TgChatID, club.Name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create club: %w", err)
	}
	return id, nil
}

func (r *clubRepository) GetByID(ctx context.Context, id int64) (*domain.Club, error) {
	query := `SELECT id, tg_chat_id, name, created_at, updated_at FROM clubs WHERE id = $1`
	var c domain.Club
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.TgChatID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("club not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get club: %w", err)
	}
	return &c, nil
}

func (r *clubRepository) GetByOwner(ctx context.Context, playerID int64) ([]*domain.Club, error) {
	query := `
		SELECT c.id, c.tg_chat_id, c.name, c.created_at, c.updated_at
		FROM clubs c
		JOIN club_members cm ON cm.club_id = c.id
		WHERE cm.player_id = $1 AND cm.role = 'owner'
		ORDER BY c.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clubs by owner: %w", err)
	}
	defer rows.Close()

	var clubs []*domain.Club
	for rows.Next() {
		var c domain.Club
		if err := rows.Scan(&c.ID, &c.TgChatID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan club: %w", err)
		}
		clubs = append(clubs, &c)
	}
	return clubs, nil
}

func (r *clubRepository) GetByTgChatID(ctx context.Context, tgChatID int64) (*domain.Club, error) {
	query := `SELECT id, tg_chat_id, name, created_at, updated_at FROM clubs WHERE tg_chat_id = $1`
	var c domain.Club
	err := r.db.Pool.QueryRow(ctx, query, tgChatID).Scan(
		&c.ID, &c.TgChatID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("club not found: tg_chat_id=%d", tgChatID)
		}
		return nil, fmt.Errorf("failed to get club by tg_chat_id: %w", err)
	}
	return &c, nil
}

func (r *clubRepository) UpdateName(ctx context.Context, id int64, name string) error {
	query := `UPDATE clubs SET name = $1 WHERE id = $2`
	tag, err := r.db.Pool.Exec(ctx, query, name, id)
	if err != nil {
		return fmt.Errorf("failed to update club name: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club not found: id=%d", id)
	}
	return nil
}

func (r *clubRepository) BindTgChatID(ctx context.Context, clubID, tgChatID int64) error {
	query := `UPDATE clubs SET tg_chat_id = $1 WHERE id = $2`
	tag, err := r.db.Pool.Exec(ctx, query, tgChatID, clubID)
	if err != nil {
		return fmt.Errorf("failed to bind tg_chat_id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club not found: id=%d", clubID)
	}
	return nil
}

func (r *clubRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM clubs WHERE id = $1`
	tag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club not found: id=%d", id)
	}
	return nil
}

// playerRepository implements domain.PlayerRepository.
type playerRepository struct {
	db *DB
}

func (r *playerRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *playerRepository) Create(ctx context.Context, player *domain.Player) (int64, error) {
	query := `
		INSERT INTO players (first_name, last_name, nickname, phone_number, email, password, tg_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query,
		player.FirstName, player.LastName, player.Nickname,
		player.PhoneNumber, player.Email, player.Password, player.TgUserID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create player: %w", err)
	}
	return id, nil
}

func (r *playerRepository) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.Player, error) {
	query := `
		SELECT id, first_name, last_name, nickname, phone_number, email, password, tg_user_id, last_seen, created_at, updated_at
		FROM players WHERE tg_user_id = $1
	`
	var p domain.Player
	err := r.db.Pool.QueryRow(ctx, query, tgUserID).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.Nickname,
		&p.PhoneNumber, &p.Email, &p.Password, &p.TgUserID,
		&p.LastSeen, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("player not found: tg_user_id=%d", tgUserID)
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return &p, nil
}

func (r *playerRepository) GetByNickname(ctx context.Context, nickname string) (*domain.Player, error) {
	query := `
		SELECT id, first_name, last_name, nickname, phone_number, email, password, tg_user_id, last_seen, created_at, updated_at
		FROM players WHERE nickname = $1
	`
	var p domain.Player
	err := r.db.Pool.QueryRow(ctx, query, nickname).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.Nickname,
		&p.PhoneNumber, &p.Email, &p.Password, &p.TgUserID,
		&p.LastSeen, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("player not found: nickname=%s", nickname)
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return &p, nil
}

func (r *playerRepository) GetByID(ctx context.Context, id int64) (*domain.Player, error) {
	query := `
		SELECT id, first_name, last_name, nickname, phone_number, email, password, tg_user_id, last_seen, created_at, updated_at
		FROM players WHERE id = $1
	`
	var p domain.Player
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.Nickname,
		&p.PhoneNumber, &p.Email, &p.Password, &p.TgUserID,
		&p.LastSeen, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("player not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return &p, nil
}

func (r *playerRepository) UpdateLastSeen(ctx context.Context, id int64) error {
	query := `UPDATE players SET last_seen = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}
	return nil
}

// clubMemberRepository implements domain.ClubMemberRepository.
type clubMemberRepository struct {
	db *DB
}

func (r *clubMemberRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *clubMemberRepository) Create(ctx context.Context, member *domain.ClubMember) (int64, error) {
	query := `
		INSERT INTO club_members (club_id, player_id, role, status, accepted)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query,
		member.ClubID, member.PlayerID, member.Role, member.Status, member.Accepted,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create club member: %w", err)
	}
	return id, nil
}

func (r *clubMemberRepository) GetByClubAndPlayer(ctx context.Context, clubID, playerID int64) (*domain.ClubMember, error) {
	query := `
		SELECT id, club_id, player_id, role, status, accepted, created_at, updated_at
		FROM club_members WHERE club_id = $1 AND player_id = $2
	`
	var m domain.ClubMember
	err := r.db.Pool.QueryRow(ctx, query, clubID, playerID).Scan(
		&m.ID, &m.ClubID, &m.PlayerID, &m.Role, &m.Status, &m.Accepted,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("club member not found: club_id=%d, player_id=%d", clubID, playerID)
		}
		return nil, fmt.Errorf("failed to get club member: %w", err)
	}
	return &m, nil
}

func (r *clubMemberRepository) GetByClubWithPlayers(ctx context.Context, clubID int64) ([]*domain.ClubMemberWithPlayer, error) {
	query := `
		SELECT cm.id, cm.club_id, cm.player_id, cm.role, cm.status, cm.accepted, cm.created_at, cm.updated_at,
		       p.id, p.first_name, p.last_name, p.nickname, p.phone_number, p.email, p.password, p.tg_user_id, p.last_seen, p.created_at, p.updated_at
		FROM club_members cm
		JOIN players p ON p.id = cm.player_id
		WHERE cm.club_id = $1
		ORDER BY cm.created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, clubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get club members: %w", err)
	}
	defer rows.Close()

	var members []*domain.ClubMemberWithPlayer
	for rows.Next() {
		var cmp domain.ClubMemberWithPlayer
		if err := rows.Scan(
			&cmp.ID, &cmp.ClubID, &cmp.PlayerID, &cmp.Role, &cmp.Status, &cmp.Accepted,
			&cmp.CreatedAt, &cmp.UpdatedAt,
			&cmp.Player.ID, &cmp.Player.FirstName, &cmp.Player.LastName, &cmp.Player.Nickname,
			&cmp.Player.PhoneNumber, &cmp.Player.Email, &cmp.Player.Password, &cmp.Player.TgUserID,
			&cmp.Player.LastSeen, &cmp.Player.CreatedAt, &cmp.Player.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan club member: %w", err)
		}
		members = append(members, &cmp)
	}
	return members, nil
}

func (r *clubMemberRepository) UpdateRole(ctx context.Context, clubID, playerID int64, role string) error {
	query := `UPDATE club_members SET role = $1 WHERE club_id = $2 AND player_id = $3`
	tag, err := r.db.Pool.Exec(ctx, query, role, clubID, playerID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club member not found: club_id=%d, player_id=%d", clubID, playerID)
	}
	return nil
}

func (r *clubMemberRepository) UpdateStatus(ctx context.Context, clubID, playerID int64, status string) error {
	query := `UPDATE club_members SET status = $1 WHERE club_id = $2 AND player_id = $3`
	tag, err := r.db.Pool.Exec(ctx, query, status, clubID, playerID)
	if err != nil {
		return fmt.Errorf("failed to update member status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club member not found: club_id=%d, player_id=%d", clubID, playerID)
	}
	return nil
}

func (r *clubMemberRepository) UpdateAccepted(ctx context.Context, clubID, playerID int64, accepted bool) error {
	query := `UPDATE club_members SET accepted = $1 WHERE club_id = $2 AND player_id = $3`
	tag, err := r.db.Pool.Exec(ctx, query, accepted, clubID, playerID)
	if err != nil {
		return fmt.Errorf("failed to update member accepted: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("club member not found: club_id=%d, player_id=%d", clubID, playerID)
	}
	return nil
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

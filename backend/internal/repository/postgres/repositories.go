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

func (r *clubRepository) GetByPlayer(ctx context.Context, playerID int64) ([]*domain.Club, error) {
	query := `
		SELECT c.id, c.tg_chat_id, c.name, c.created_at, c.updated_at
		FROM clubs c
		JOIN club_members cm ON cm.club_id = c.id
		WHERE cm.player_id = $1
		ORDER BY c.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clubs by player: %w", err)
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

func (r *gameRepository) Create(ctx context.Context, game *domain.Game) (int64, error) {
	query := `
		INSERT INTO games (
			club_id, banker_id, game_type, currency, money_model,
			chip_value, buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys,
			duration, start_time, end_time, status, min_players, max_players,
			ranking_primary, ranking_secondary
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id
	`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query,
		game.ClubID, game.BankerID, game.GameType, game.Currency, game.MoneyModel,
		game.ChipValue, game.BuyInAmount, game.RebuyAllowed, game.RebuyPrice, game.MaxRebuys,
		game.Duration, game.StartTime, game.EndTime, game.Status, game.MinPlayers, game.MaxPlayers,
		game.RankingPrimary, game.RankingSecondary,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create game: %w", err)
	}
	return id, nil
}

func (r *gameRepository) GetByID(ctx context.Context, id int64) (*domain.Game, error) {
	query := `
		SELECT id, club_id, banker_id, game_type, currency, money_model,
			chip_value, buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys,
			duration, start_time, end_time, status, min_players, max_players,
			ranking_primary, ranking_secondary, created_at, updated_at
		FROM games WHERE id = $1
	`
	var g domain.Game
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.ClubID, &g.BankerID, &g.GameType, &g.Currency, &g.MoneyModel,
		&g.ChipValue, &g.BuyInAmount, &g.RebuyAllowed, &g.RebuyPrice, &g.MaxRebuys,
		&g.Duration, &g.StartTime, &g.EndTime, &g.Status, &g.MinPlayers, &g.MaxPlayers,
		&g.RankingPrimary, &g.RankingSecondary, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("game not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	return &g, nil
}

func (r *gameRepository) GetByClub(ctx context.Context, clubID int64) ([]*domain.Game, error) {
	query := `
		SELECT id, club_id, banker_id, game_type, currency, money_model,
			chip_value, buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys,
			duration, start_time, end_time, status, min_players, max_players,
			ranking_primary, ranking_secondary, created_at, updated_at
		FROM games WHERE club_id = $1
		ORDER BY start_time DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, clubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by club: %w", err)
	}
	defer rows.Close()

	var games []*domain.Game
	for rows.Next() {
		var g domain.Game
		if err := rows.Scan(
			&g.ID, &g.ClubID, &g.BankerID, &g.GameType, &g.Currency, &g.MoneyModel,
			&g.ChipValue, &g.BuyInAmount, &g.RebuyAllowed, &g.RebuyPrice, &g.MaxRebuys,
			&g.Duration, &g.StartTime, &g.EndTime, &g.Status, &g.MinPlayers, &g.MaxPlayers,
			&g.RankingPrimary, &g.RankingSecondary, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &g)
	}
	return games, nil
}

func (r *gameRepository) GetActiveByClub(ctx context.Context, clubID int64) (*domain.Game, error) {
	query := `
		SELECT id, club_id, banker_id, game_type, currency, money_model,
			chip_value, buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys,
			duration, start_time, end_time, status, min_players, max_players,
			ranking_primary, ranking_secondary, created_at, updated_at
		FROM games WHERE club_id = $1 AND status = 'active'
		ORDER BY start_time DESC
		LIMIT 1
	`
	var g domain.Game
	err := r.db.Pool.QueryRow(ctx, query, clubID).Scan(
		&g.ID, &g.ClubID, &g.BankerID, &g.GameType, &g.Currency, &g.MoneyModel,
		&g.ChipValue, &g.BuyInAmount, &g.RebuyAllowed, &g.RebuyPrice, &g.MaxRebuys,
		&g.Duration, &g.StartTime, &g.EndTime, &g.Status, &g.MinPlayers, &g.MaxPlayers,
		&g.RankingPrimary, &g.RankingSecondary, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no active game found for club: club_id=%d", clubID)
		}
		return nil, fmt.Errorf("failed to get active game: %w", err)
	}
	return &g, nil
}

func (r *gameRepository) Update(ctx context.Context, game *domain.Game) error {
	query := `
		UPDATE games SET
			banker_id = $1, game_type = $2, currency = $3, money_model = $4,
			chip_value = $5, buy_in_amount = $6, rebuy_allowed = $7, rebuy_price = $8,
			max_rebuys = $9, duration = $10, start_time = $11, min_players = $12,
			max_players = $13, ranking_primary = $14, ranking_secondary = $15
		WHERE id = $16
	`
	tag, err := r.db.Pool.Exec(ctx, query,
		game.BankerID, game.GameType, game.Currency, game.MoneyModel,
		game.ChipValue, game.BuyInAmount, game.RebuyAllowed, game.RebuyPrice,
		game.MaxRebuys, game.Duration, game.StartTime, game.MinPlayers,
		game.MaxPlayers, game.RankingPrimary, game.RankingSecondary, game.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("game not found: id=%d", game.ID)
	}
	return nil
}

func (r *gameRepository) Cancel(ctx context.Context, id int64) error {
	query := `UPDATE games SET status = 'cancelled' WHERE id = $1`
	tag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel game: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("game not found: id=%d", id)
	}
	return nil
}

// gameParticipantRepository implements domain.GameParticipantRepository.
type gameParticipantRepository struct {
	db *DB
}

func (r *gameParticipantRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *gameParticipantRepository) Create(ctx context.Context, participant *domain.GameParticipant) (int64, error) {
	query := `
		INSERT INTO game_participants (game_id, player_id, buy_in_count, rebuy_count, chips_end, payout_amount, place, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query,
		participant.GameID, participant.PlayerID, participant.BuyInCount,
		participant.RebuyCount, participant.ChipsEnd, participant.PayoutAmount,
		participant.Place, participant.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create game participant: %w", err)
	}
	return id, nil
}

func (r *gameParticipantRepository) GetByID(ctx context.Context, id int64) (*domain.GameParticipant, error) {
	query := `
		SELECT id, game_id, player_id, buy_in_count, rebuy_count, chips_end, payout_amount, place, status, created_at, updated_at
		FROM game_participants WHERE id = $1
	`
	var p domain.GameParticipant
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.GameID, &p.PlayerID, &p.BuyInCount, &p.RebuyCount,
		&p.ChipsEnd, &p.PayoutAmount, &p.Place, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("game participant not found: id=%d", id)
		}
		return nil, fmt.Errorf("failed to get game participant: %w", err)
	}
	return &p, nil
}

func (r *gameParticipantRepository) GetByGame(ctx context.Context, gameID int64) ([]*domain.GameParticipant, error) {
	query := `
		SELECT id, game_id, player_id, buy_in_count, rebuy_count, chips_end, payout_amount, place, status, created_at, updated_at
		FROM game_participants WHERE game_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game participants: %w", err)
	}
	defer rows.Close()

	var participants []*domain.GameParticipant
	for rows.Next() {
		var p domain.GameParticipant
		if err := rows.Scan(
			&p.ID, &p.GameID, &p.PlayerID, &p.BuyInCount, &p.RebuyCount,
			&p.ChipsEnd, &p.PayoutAmount, &p.Place, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan game participant: %w", err)
		}
		participants = append(participants, &p)
	}
	return participants, nil
}

func (r *gameParticipantRepository) GetByGameWithPlayers(ctx context.Context, gameID int64) ([]*domain.GameParticipantWithPlayer, error) {
	query := `
		SELECT gp.id, gp.game_id, gp.player_id, gp.buy_in_count, gp.rebuy_count,
		       gp.chips_end, gp.payout_amount, gp.place, gp.status, gp.created_at, gp.updated_at,
		       p.id, p.first_name, p.last_name, p.nickname, p.phone_number, p.email, p.password,
		       p.tg_user_id, p.last_seen, p.created_at, p.updated_at
		FROM game_participants gp
		JOIN players p ON p.id = gp.player_id
		WHERE gp.game_id = $1
		ORDER BY gp.id ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game participants with players: %w", err)
	}
	defer rows.Close()

	var participants []*domain.GameParticipantWithPlayer
	for rows.Next() {
		var gp domain.GameParticipantWithPlayer
		if err := rows.Scan(
			&gp.ID, &gp.GameID, &gp.PlayerID, &gp.BuyInCount, &gp.RebuyCount,
			&gp.ChipsEnd, &gp.PayoutAmount, &gp.Place, &gp.Status, &gp.CreatedAt, &gp.UpdatedAt,
			&gp.Player.ID, &gp.Player.FirstName, &gp.Player.LastName, &gp.Player.Nickname,
			&gp.Player.PhoneNumber, &gp.Player.Email, &gp.Player.Password,
			&gp.Player.TgUserID, &gp.Player.LastSeen, &gp.Player.CreatedAt, &gp.Player.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan game participant with player: %w", err)
		}
		participants = append(participants, &gp)
	}
	return participants, nil
}

func (r *gameParticipantRepository) GetByGameAndPlayer(ctx context.Context, gameID, playerID int64) (*domain.GameParticipant, error) {
	query := `
		SELECT id, game_id, player_id, buy_in_count, rebuy_count, chips_end, payout_amount, place, status, created_at, updated_at
		FROM game_participants WHERE game_id = $1 AND player_id = $2
	`
	var p domain.GameParticipant
	err := r.db.Pool.QueryRow(ctx, query, gameID, playerID).Scan(
		&p.ID, &p.GameID, &p.PlayerID, &p.BuyInCount, &p.RebuyCount,
		&p.ChipsEnd, &p.PayoutAmount, &p.Place, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("game participant not found: game_id=%d, player_id=%d", gameID, playerID)
		}
		return nil, fmt.Errorf("failed to get game participant: %w", err)
	}
	return &p, nil
}

func (r *gameParticipantRepository) Update(ctx context.Context, participant *domain.GameParticipant) error {
	query := `
		UPDATE game_participants SET
			buy_in_count = $1, rebuy_count = $2, chips_end = $3, payout_amount = $4,
			place = $5, status = $6
		WHERE id = $7
	`
	tag, err := r.db.Pool.Exec(ctx, query,
		participant.BuyInCount, participant.RebuyCount, participant.ChipsEnd,
		participant.PayoutAmount, participant.Place, participant.Status, participant.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update game participant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("game participant not found: id=%d", participant.ID)
	}
	return nil
}

func (r *gameParticipantRepository) UpdateStatus(ctx context.Context, gameID, playerID int64, status string) error {
	query := `UPDATE game_participants SET status = $1 WHERE game_id = $2 AND player_id = $3`
	tag, err := r.db.Pool.Exec(ctx, query, status, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to update game participant status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("game participant not found: game_id=%d, player_id=%d", gameID, playerID)
	}
	return nil
}

func (r *gameParticipantRepository) Delete(ctx context.Context, gameID, playerID int64) error {
	query := `DELETE FROM game_participants WHERE game_id = $1 AND player_id = $2`
	tag, err := r.db.Pool.Exec(ctx, query, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to delete game participant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("game participant not found: game_id=%d, player_id=%d", gameID, playerID)
	}
	return nil
}

// eventRepository implements domain.EventRepository.
type eventRepository struct {
	db *DB
}

func (r *eventRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *eventRepository) Create(ctx context.Context, event *domain.Event) (int64, error) {
	query := `
		INSERT INTO events (game_id, player_id, type, old_value, new_value, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err := r.db.Pool.QueryRow(ctx, query,
		event.GameID, event.PlayerID, event.Type, event.OldValue, event.NewValue,
		event.Metadata, event.CreatedBy,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}
	return id, nil
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

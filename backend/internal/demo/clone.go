package demo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type cloneResult struct {
	ClubID   int64
	PlayerID int64 // visitor = owner (Thor)
	TgUserID int64
}

// cloneTemplateClub deep-copies the Asgard template club into a personal session club.
// Unique player fields are remapped using demo_clone_seq so concurrent clones do not collide.
//
// Important: all SELECT rows on tx must be fully consumed and closed before any further
// Query/Exec on the same tx — pgx allows only one concurrent query per connection.
func (s *Service) cloneTemplateClub(ctx context.Context, tx pgx.Tx, templateClubID int64) (*cloneResult, error) {
	var cloneN int64
	if err := tx.QueryRow(ctx, `SELECT nextval('demo_clone_seq')`).Scan(&cloneN); err != nil {
		return nil, fmt.Errorf("allocate clone sequence: %w", err)
	}

	var newClubID int64
	err := tx.QueryRow(ctx, `
		INSERT INTO clubs (tg_chat_id, name, is_template)
		SELECT NULL, name, FALSE
		FROM clubs WHERE id = $1
		RETURNING id
	`, templateClubID).Scan(&newClubID)
	if err != nil {
		return nil, fmt.Errorf("clone club: %w", err)
	}

	type tmplPlayer struct {
		ID         int64
		FirstName  string
		LastName   string
		Nickname   *string
		Phone      *string
		Email      *string
		Password   *string
		TgUserID   *int64
		TgUserName *string
	}

	rows, err := tx.Query(ctx, `
		SELECT p.id, p.first_name, p.last_name, p.nickname, p.phone_number, p.email, p.password,
		       p.tg_user_id, p.tg_user_name
		FROM players p
		JOIN club_members cm ON cm.player_id = p.id
		WHERE cm.club_id = $1
		ORDER BY p.id
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template players: %w", err)
	}

	var players []tmplPlayer
	for rows.Next() {
		var p tmplPlayer
		if err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.Nickname, &p.Phone, &p.Email, &p.Password,
			&p.TgUserID, &p.TgUserName,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan template player: %w", err)
		}
		players = append(players, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	playerMap := make(map[int64]int64, len(players))
	for i, p := range players {
		newPhone := uniquePhone(p.Phone, cloneN, i)
		newEmail := uniqueEmail(p.Email, cloneN, i)
		newTgID := uniqueTgUserID(p.TgUserID, cloneN, i)
		newTgName := uniqueTgUserName(p.TgUserName, cloneN, i)

		var newPlayerID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO players (
				first_name, last_name, nickname, phone_number, email, password,
				tg_user_id, tg_user_name
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id
		`, p.FirstName, p.LastName, p.Nickname, newPhone, newEmail, p.Password, newTgID, newTgName,
		).Scan(&newPlayerID)
		if err != nil {
			return nil, fmt.Errorf("clone player %d: %w", p.ID, err)
		}
		playerMap[p.ID] = newPlayerID
	}

	type tmplMember struct {
		ID       int64
		PlayerID int64
		Role     string
		Status   string
		Accepted bool
	}
	mrows, err := tx.Query(ctx, `
		SELECT id, player_id, role, status, accepted
		FROM club_members WHERE club_id = $1 ORDER BY id
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template members: %w", err)
	}

	var members []tmplMember
	for mrows.Next() {
		var m tmplMember
		if err := mrows.Scan(&m.ID, &m.PlayerID, &m.Role, &m.Status, &m.Accepted); err != nil {
			mrows.Close()
			return nil, fmt.Errorf("scan template member: %w", err)
		}
		members = append(members, m)
	}
	if err := mrows.Err(); err != nil {
		mrows.Close()
		return nil, err
	}
	mrows.Close()

	memberMap := make(map[int64]int64, len(members))
	var ownerPlayerID, ownerTgUserID int64
	for _, m := range members {
		newPlayerID, ok := playerMap[m.PlayerID]
		if !ok {
			return nil, fmt.Errorf("missing player map for %d", m.PlayerID)
		}
		var newMemberID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO club_members (club_id, player_id, role, status, accepted)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, newClubID, newPlayerID, m.Role, m.Status, m.Accepted).Scan(&newMemberID)
		if err != nil {
			return nil, fmt.Errorf("clone member %d: %w", m.ID, err)
		}
		memberMap[m.ID] = newMemberID
		if m.Role == "owner" {
			ownerPlayerID = newPlayerID
			if err := tx.QueryRow(ctx, `SELECT tg_user_id FROM players WHERE id = $1`, newPlayerID).Scan(&ownerTgUserID); err != nil {
				return nil, fmt.Errorf("load owner tg_user_id: %w", err)
			}
		}
	}
	if ownerPlayerID == 0 {
		return nil, fmt.Errorf("template club %d has no owner", templateClubID)
	}

	type tmplGame struct {
		ID                  int64
		BankerID            int64
		GameType            string
		Currency            string
		MoneyModel          string
		ChipValue           float64
		BuyInAmount         float64
		RebuyAllowed        bool
		RebuyPrice          *float64
		MaxRebuys           *int
		Duration            *string
		StartTime           interface{}
		EndTime             interface{}
		Status              string
		MinPlayers          int
		MaxPlayers          int
		RankingPrimary      string
		RankingSecondary    *string
		TimerPausedAt       interface{}
		TimerPausedDuration string
		TimerNotified       bool
	}

	grows, err := tx.Query(ctx, `
		SELECT id, banker_id, game_type, currency, money_model, chip_value,
		       buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys,
		       duration::text, start_time, end_time, status, min_players, max_players,
		       ranking_primary, ranking_secondary,
		       timer_paused_at, timer_paused_duration::text, timer_notified
		FROM games WHERE club_id = $1 ORDER BY id
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template games: %w", err)
	}

	var games []tmplGame
	for grows.Next() {
		var g tmplGame
		if err := grows.Scan(
			&g.ID, &g.BankerID, &g.GameType, &g.Currency, &g.MoneyModel, &g.ChipValue,
			&g.BuyInAmount, &g.RebuyAllowed, &g.RebuyPrice, &g.MaxRebuys,
			&g.Duration, &g.StartTime, &g.EndTime, &g.Status, &g.MinPlayers, &g.MaxPlayers,
			&g.RankingPrimary, &g.RankingSecondary,
			&g.TimerPausedAt, &g.TimerPausedDuration, &g.TimerNotified,
		); err != nil {
			grows.Close()
			return nil, fmt.Errorf("scan template game: %w", err)
		}
		games = append(games, g)
	}
	if err := grows.Err(); err != nil {
		grows.Close()
		return nil, err
	}
	grows.Close()

	gameMap := make(map[int64]int64, len(games))
	for _, g := range games {
		newBankerID, ok := memberMap[g.BankerID]
		if !ok {
			return nil, fmt.Errorf("missing banker member map for %d", g.BankerID)
		}
		var newGameID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO games (
				club_id, banker_id, game_type, currency, money_model, chip_value,
				buy_in_amount, rebuy_allowed, rebuy_price, max_rebuys, duration,
				start_time, end_time, status, min_players, max_players,
				ranking_primary, ranking_secondary,
				timer_paused_at, timer_paused_duration, timer_notified
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11::interval,
				$12, $13, $14, $15, $16,
				$17, $18,
				$19, $20::interval, $21
			) RETURNING id
		`, newClubID, newBankerID, g.GameType, g.Currency, g.MoneyModel, g.ChipValue,
			g.BuyInAmount, g.RebuyAllowed, g.RebuyPrice, g.MaxRebuys, g.Duration,
			g.StartTime, g.EndTime, g.Status, g.MinPlayers, g.MaxPlayers,
			g.RankingPrimary, g.RankingSecondary,
			g.TimerPausedAt, g.TimerPausedDuration, g.TimerNotified,
		).Scan(&newGameID)
		if err != nil {
			return nil, fmt.Errorf("clone game %d: %w", g.ID, err)
		}
		gameMap[g.ID] = newGameID
	}

	type tmplParticipant struct {
		GameID   int64
		PlayerID int64
		BuyIn    int
		Rebuy    int
		ChipsEnd *float64
		Payout   *float64
		Place    *int
		Status   string
	}

	prows, err := tx.Query(ctx, `
		SELECT gp.game_id, gp.player_id, gp.buy_in_count, gp.rebuy_count,
		       gp.chips_end, gp.payout_amount, gp.place, gp.status
		FROM game_participants gp
		JOIN games g ON g.id = gp.game_id
		WHERE g.club_id = $1
		ORDER BY gp.id
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template participants: %w", err)
	}

	var participants []tmplParticipant
	for prows.Next() {
		var p tmplParticipant
		if err := prows.Scan(
			&p.GameID, &p.PlayerID, &p.BuyIn, &p.Rebuy,
			&p.ChipsEnd, &p.Payout, &p.Place, &p.Status,
		); err != nil {
			prows.Close()
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		participants = append(participants, p)
	}
	if err := prows.Err(); err != nil {
		prows.Close()
		return nil, err
	}
	prows.Close()

	for _, p := range participants {
		if _, err := tx.Exec(ctx, `
			INSERT INTO game_participants (
				game_id, player_id, buy_in_count, rebuy_count,
				chips_end, payout_amount, place, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, gameMap[p.GameID], playerMap[p.PlayerID], p.BuyIn, p.Rebuy, p.ChipsEnd, p.Payout, p.Place, p.Status); err != nil {
			return nil, fmt.Errorf("clone participant: %w", err)
		}
	}

	type tmplEvent struct {
		GameID    int64
		PlayerID  int64
		Type      string
		OldVal    *float64
		NewVal    *float64
		Metadata  []byte
		CreatedAt interface{}
		CreatedBy int64
	}

	erows, err := tx.Query(ctx, `
		SELECT e.game_id, e.player_id, e.type, e.old_value, e.new_value, e.metadata, e.created_at, e.created_by
		FROM events e
		JOIN games g ON g.id = e.game_id
		WHERE g.club_id = $1
		ORDER BY e.id
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template events: %w", err)
	}

	var events []tmplEvent
	for erows.Next() {
		var e tmplEvent
		if err := erows.Scan(
			&e.GameID, &e.PlayerID, &e.Type, &e.OldVal, &e.NewVal,
			&e.Metadata, &e.CreatedAt, &e.CreatedBy,
		); err != nil {
			erows.Close()
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	if err := erows.Err(); err != nil {
		erows.Close()
		return nil, err
	}
	erows.Close()

	for _, e := range events {
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (game_id, player_id, type, old_value, new_value, metadata, created_at, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, gameMap[e.GameID], playerMap[e.PlayerID], e.Type, e.OldVal, e.NewVal, e.Metadata, e.CreatedAt, memberMap[e.CreatedBy]); err != nil {
			return nil, fmt.Errorf("clone event: %w", err)
		}
	}

	type tmplStats struct {
		PlayerID      int64
		TotalGames    int
		TotalBuyIn    float64
		TotalRebuy    float64
		TotalRebuys   int
		TotalInvested float64
		TotalChips    float64
		TotalProfit   float64
		BiggestWin    float64
		BiggestLoss   float64
		GamesWon      int
		Podiums       int
		ROI           float64
		ITM           float64
	}

	srows, err := tx.Query(ctx, `
		SELECT player_id, total_games, total_buy_in_amount, total_rebuy_amount, total_rebuys_count,
		       total_invested, total_chips, total_profit, biggest_win, biggest_loss,
		       games_won, podiums, roi, itm
		FROM player_statistics WHERE club_id = $1
	`, templateClubID)
	if err != nil {
		return nil, fmt.Errorf("load template statistics: %w", err)
	}

	var stats []tmplStats
	for srows.Next() {
		var st tmplStats
		if err := srows.Scan(
			&st.PlayerID, &st.TotalGames, &st.TotalBuyIn, &st.TotalRebuy, &st.TotalRebuys,
			&st.TotalInvested, &st.TotalChips, &st.TotalProfit, &st.BiggestWin, &st.BiggestLoss,
			&st.GamesWon, &st.Podiums, &st.ROI, &st.ITM,
		); err != nil {
			srows.Close()
			return nil, fmt.Errorf("scan statistics: %w", err)
		}
		stats = append(stats, st)
	}
	if err := srows.Err(); err != nil {
		srows.Close()
		return nil, err
	}
	srows.Close()

	for _, st := range stats {
		if _, err := tx.Exec(ctx, `
			INSERT INTO player_statistics (
				player_id, club_id,
				total_games, total_buy_in_amount, total_rebuy_amount, total_rebuys_count,
				total_invested, total_chips, total_profit,
				biggest_win, biggest_loss, games_won, podiums, roi, itm
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`, playerMap[st.PlayerID], newClubID, st.TotalGames, st.TotalBuyIn, st.TotalRebuy, st.TotalRebuys,
			st.TotalInvested, st.TotalChips, st.TotalProfit, st.BiggestWin, st.BiggestLoss, st.GamesWon, st.Podiums, st.ROI, st.ITM,
		); err != nil {
			return nil, fmt.Errorf("clone statistics: %w", err)
		}
	}

	return &cloneResult{
		ClubID:   newClubID,
		PlayerID: ownerPlayerID,
		TgUserID: ownerTgUserID,
	}, nil
}

func uniquePhone(base *string, cloneN int64, idx int) *string {
	if base == nil || *base == "" {
		return nil
	}
	v := fmt.Sprintf("%s_s%d_%d", *base, cloneN, idx)
	return &v
}

func uniqueEmail(base *string, cloneN int64, idx int) *string {
	if base == nil || *base == "" {
		v := fmt.Sprintf("demo_s%d_%d@hpclub.com", cloneN, idx)
		return &v
	}
	local, domain, ok := strings.Cut(*base, "@")
	if !ok {
		v := fmt.Sprintf("%s+s%d_%d@hpclub.com", *base, cloneN, idx)
		return &v
	}
	v := fmt.Sprintf("%s+s%d_%d@%s", local, cloneN, idx, domain)
	return &v
}

func uniqueTgUserID(base *int64, cloneN int64, idx int) *int64 {
	// Synthetic positive IDs outside the template range (9999999999980+).
	id := int64(1_100_000_000_000) + cloneN*1000 + int64(idx)
	if base != nil {
		// Preserve relative ordering within the club while staying unique per clone.
		id = *base - 9_999_999_999_980 + 1_100_000_000_000 + cloneN*1000
	}
	return &id
}

func uniqueTgUserName(base *string, cloneN int64, idx int) *string {
	if base == nil || *base == "" {
		v := fmt.Sprintf("demo_s%d_%d", cloneN, idx)
		return &v
	}
	v := fmt.Sprintf("%s_s%d", *base, cloneN)
	return &v
}

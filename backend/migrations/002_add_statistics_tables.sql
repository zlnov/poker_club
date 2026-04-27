-- Migration: Add statistics tables for caching aggregates
-- Created: 2026-04-15

-- Player statistics table (cached aggregates per player per club)
CREATE TABLE IF NOT EXISTS player_statistics (
    id          BIGSERIAL PRIMARY KEY,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    total_games INTEGER NOT NULL DEFAULT 0,
    total_buy_in NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_rebuy NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_invested NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_chips NUMERIC(12,2) NOT NULL DEFAULT 0,
    profit      NUMERIC(12,2) NOT NULL DEFAULT 0,
    roi         NUMERIC(8,4) NOT NULL DEFAULT 0, -- percentage
    itm         NUMERIC(8,4) NOT NULL DEFAULT 0, -- percentage
    win_rate    NUMERIC(8,4) NOT NULL DEFAULT 0, -- percentage
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (player_id, club_id)
);

-- Indexes for player statistics
CREATE INDEX IF NOT EXISTS idx_player_statistics_player_id ON player_statistics(player_id);
CREATE INDEX IF NOT EXISTS idx_player_statistics_club_id ON player_statistics(club_id);

-- Optional: Game statistics table (cached aggregates per game)
-- This is optional because game aggregates can be computed from participants and games quickly.
-- But if we want to cache for leaderboards or frequent queries, we can add:
CREATE TABLE IF NOT EXISTS game_statistics (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    total_invested NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_chips NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (game_id)
);

-- Index for game statistics
CREATE INDEX IF NOT EXISTS idx_game_statistics_game_id ON game_statistics(game_id);

-- Note: We do not add new columns for results or rebuys because:
--   - Results (chips_end, place) are already in game_participants.
--   - Rebuys are tracked via rebuy_count in game_participants and detailed in events table (type='rebuy').
--   - If we wanted to store each rebuy with timestamp, we already have events table.

-- However, if we want to store each rebuy as a separate row for audit, we already have events.
-- So no new tables needed for rebuys.

-- For completeness, we could add a column to games to store the prize places (e.g., jsonb) to determine ITM precisely.
-- But for simplicity, we assume top 3 are ITM in the statistics calculation.

-- Add prize_places column to games if needed (optional)
-- ALTER TABLE games ADD COLUMN IF NOT EXISTS prize_places JSONB; -- e.g., [1,2,3] for top 3 get prize
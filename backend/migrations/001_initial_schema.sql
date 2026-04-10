-- Migration: Initial schema
-- Created: 2024-01-01

-- Clubs table
CREATE TABLE IF NOT EXISTS clubs (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id          BIGSERIAL PRIMARY KEY,
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    nickname    TEXT,
    phone_number TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Club members table
CREATE TABLE IF NOT EXISTS club_members (
    id          BIGSERIAL PRIMARY KEY,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('admin','member')),
    status      TEXT NOT NULL CHECK (status IN ('pending','active','banned')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (club_id, player_id)
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id                  BIGSERIAL PRIMARY KEY,
    club_id             BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    banker_id           BIGINT NOT NULL REFERENCES club_members(id) ON DELETE RESTRICT,
    type                TEXT NOT NULL CHECK (type IN ('cash_time','cash_open','tournament')),
    money_model         TEXT NOT NULL DEFAULT 'real',
    buy_in_amount       NUMERIC(12,2) NOT NULL,
    rebuy_allowed       BOOLEAN NOT NULL DEFAULT FALSE,
    rebuy_amount        NUMERIC(12,2),
    max_rebuys_per_player INTEGER,
    duration            INTERVAL,
    start_time          TIMESTAMPTZ NOT NULL,
    end_time            TIMESTAMPTZ,
    min_players         INTEGER NOT NULL,
    max_players         INTEGER NOT NULL,
    ranking_primary     TEXT NOT NULL,
    ranking_secondary   TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Game participants table
CREATE TABLE IF NOT EXISTS game_participants (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id       BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    buy_in_count    INTEGER NOT NULL DEFAULT 0,
    rebuy_count     INTEGER NOT NULL DEFAULT 0,
    chips_end       NUMERIC(12,2),
    place           INTEGER,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, player_id)
);

-- Events table (Event Log)
CREATE TABLE IF NOT EXISTS events (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type        TEXT NOT NULL CHECK (type IN ('buy_in','rebuy','chips_set','correction')),
    amount      NUMERIC(12,2),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  BIGINT NOT NULL REFERENCES club_members(id) ON DELETE SET NULL
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_events_player_id ON events(player_id);
CREATE INDEX IF NOT EXISTS idx_events_game_id ON events(game_id);
CREATE INDEX IF NOT EXISTS idx_game_participants_game_id ON game_participants(game_id);
CREATE INDEX IF NOT EXISTS idx_game_participants_player_id ON game_participants(player_id);
CREATE INDEX IF NOT EXISTS idx_club_members_club_id ON club_members(club_id);
CREATE INDEX IF NOT EXISTS idx_club_members_player_id ON club_members(player_id);
CREATE INDEX IF NOT EXISTS idx_games_club_id ON games(club_id);
CREATE INDEX IF NOT EXISTS idx_games_banker_id ON games(banker_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_clubs_updated_at BEFORE UPDATE ON clubs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_club_members_updated_at BEFORE UPDATE ON club_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_participants_updated_at BEFORE UPDATE ON game_participants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

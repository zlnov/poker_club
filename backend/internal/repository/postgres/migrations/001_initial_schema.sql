-- +goose Up
-- +goose StatementBegin

-- Clubs table
CREATE TABLE IF NOT EXISTS clubs (
    id          BIGSERIAL PRIMARY KEY,
    tg_chat_id  BIGINT UNIQUE,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id          BIGSERIAL PRIMARY KEY,
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    nickname    TEXT NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    tg_user_id  BIGINT UNIQUE,
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Club members table
CREATE TABLE IF NOT EXISTS club_members (
    id          BIGSERIAL PRIMARY KEY,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('owner','admin','member')),
    status      TEXT NOT NULL CHECK (status IN ('pending','active','banned','left')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (club_id, player_id)
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id                  BIGSERIAL PRIMARY KEY,
    club_id             BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    banker_id           BIGINT NOT NULL REFERENCES club_members(id) ON DELETE RESTRICT,
    game_type           TEXT NOT NULL CHECK (game_type IN ('cash','tournament')),
    currency            CHAR(3) NOT NULL CHECK (currency IN ('USD', 'EUR', 'RUB', 'KZT','BYN')) DEFAULT 'RUB',
    money_model         TEXT NOT NULL CHECK (money_model IN ('real', 'points', 'virtual', 'practice')) DEFAULT 'real',
    chip_value          NUMERIC(12,2) NOT NULL DEFAULT 1.0,
    buy_in_amount       NUMERIC(12,2) NOT NULL,
    rebuy_allowed       BOOLEAN NOT NULL DEFAULT FALSE,
    rebuy_price         NUMERIC(12,2),
    max_rebuys          INTEGER,
    duration            INTERVAL,
    start_time          TIMESTAMPTZ NOT NULL,
    end_time            TIMESTAMPTZ,
    status              TEXT NOT NULL CHECK (status IN ('planned', 'active','finished','cancelled')) DEFAULT 'planned',
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
    payout_amount   NUMERIC(12,2),
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
    old_value   NUMERIC(12,2),
    new_value   NUMERIC(12,2),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  BIGINT NOT NULL REFERENCES club_members(id) ON DELETE RESTRICT
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_events_game_created_at ON events(game_id, created_at);
CREATE INDEX IF NOT EXISTS idx_events_player_id ON events(player_id);
CREATE INDEX IF NOT EXISTS idx_game_participants_player_id ON game_participants(player_id);
CREATE INDEX IF NOT EXISTS idx_club_members_player_id ON club_members(player_id);
CREATE INDEX IF NOT EXISTS idx_games_club_start_time ON games(club_id, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_games_banker_id ON games(banker_id);

-- Player statistics table (cached aggregates per player per club)
CREATE TABLE IF NOT EXISTS player_statistics (
    id          BIGSERIAL PRIMARY KEY,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    total_games INTEGER NOT NULL DEFAULT 0,
    total_buy_in_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_rebuy_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_rebuys_count INTEGER NOT NULL DEFAULT 0,
    total_invested NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_chips NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_profit NUMERIC(12,2) NOT NULL DEFAULT 0,
    biggest_win NUMERIC(12,2) NOT NULL DEFAULT 0,
    biggest_loss NUMERIC(12,2) NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,
    podiums INTEGER NOT NULL DEFAULT 0,
    roi NUMERIC(8,4) NOT NULL DEFAULT 0,
    itm NUMERIC(8,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (player_id, club_id)
);

-- Indexes for player statistics
CREATE INDEX IF NOT EXISTS idx_player_statistics_club_id ON player_statistics(club_id);

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
CREATE TRIGGER update_player_statistics_updated_at BEFORE UPDATE ON player_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS player_statistics;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS game_participants;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS club_members;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS clubs;

-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin

-- Demo sessions track opaque bearer tokens for /api/v1/demo/* access.
-- The table is part of the shared Goose schema so demo DB migrations stay aligned
-- with production. Production traffic never uses this table.
CREATE TABLE IF NOT EXISTS demo_sessions (
    id          BIGSERIAL PRIMARY KEY,
    token_hash  TEXT NOT NULL UNIQUE,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_demo_sessions_expires_at ON demo_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_demo_sessions_club_id ON demo_sessions(club_id);

-- Sequence used by the Asgard demo seed script for unique suffixes.
CREATE SEQUENCE IF NOT EXISTS demo_seed_seq START 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS demo_sessions;
DROP SEQUENCE IF EXISTS demo_seed_seq;

-- +goose StatementEnd

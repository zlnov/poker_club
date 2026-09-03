-- +goose Up
-- +goose StatementBegin

-- Refresh tokens table for JWT refresh token storage.
-- Access JWT is stateless (not stored in DB).
-- Refresh tokens are stored as hashes (SHA-256) only.
-- Raw refresh token is never stored.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          BIGSERIAL PRIMARY KEY,
    player_id   BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at  TIMESTAMPTZ
);

-- Index for efficient lookup by player (for listing/revoking user's tokens).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_player_id ON refresh_tokens(player_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS refresh_tokens;

-- +goose StatementEnd

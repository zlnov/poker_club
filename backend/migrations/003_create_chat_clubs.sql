-- Migration: create_chat_clubs
-- This migration creates a table that maps a Telegram chat ID to a club ID.
-- It is used by the bot to resolve which club a command originates from.

CREATE TABLE IF NOT EXISTS chat_clubs (
    id          BIGSERIAL PRIMARY KEY,
    chat_id     BIGINT NOT NULL UNIQUE,
    club_id     BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Index for quick lookup by chat_id
CREATE INDEX IF NOT EXISTS idx_chat_clubs_chat_id ON chat_clubs(chat_id);


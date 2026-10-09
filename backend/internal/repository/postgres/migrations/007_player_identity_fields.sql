-- +goose Up
-- +goose StatementBegin

-- Separate Poker Club nickname from Telegram username.
-- Make password nullable so Telegram-only players can exist without a password.

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS tg_user_name TEXT NULL;

ALTER TABLE players
    ALTER COLUMN nickname DROP NOT NULL,
    ALTER COLUMN password DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback requires no NULL nicknames/passwords.
UPDATE players SET nickname = COALESCE(nickname, '') WHERE nickname IS NULL;
UPDATE players SET password = COALESCE(password, '') WHERE password IS NULL;

ALTER TABLE players
    ALTER COLUMN nickname SET NOT NULL,
    ALTER COLUMN password SET NOT NULL;

ALTER TABLE players
    DROP COLUMN IF EXISTS tg_user_name;

-- +goose StatementEnd

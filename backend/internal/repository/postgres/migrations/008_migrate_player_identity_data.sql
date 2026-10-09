-- +goose Up
-- +goose StatementBegin

-- Convert empty password strings to NULL (Telegram-created players).
-- Existing bcrypt hashes are preserved.
UPDATE players
SET password = NULL
WHERE password = '';

-- Split historical nickname semantics for Telegram-linked players:
-- - Real Telegram username for existing users → nickname.
-- - nickname remains unchanged.
UPDATE players
SET tg_user_name = nickname
WHERE tg_user_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Best-effort reverse: restore nickname from tg_user_name where present.
-- Cannot restore firstName-fallback nicknames or empty passwords accurately.
UPDATE players
SET nickname = tg_user_name
WHERE tg_user_id IS NOT NULL
  AND tg_user_name IS NOT NULL
  AND nickname IS NULL;

UPDATE players
SET password = ''
WHERE password IS NULL
  AND tg_user_id IS NOT NULL;

UPDATE players
SET tg_user_name = NULL
WHERE tg_user_id IS NOT NULL;

-- +goose StatementEnd

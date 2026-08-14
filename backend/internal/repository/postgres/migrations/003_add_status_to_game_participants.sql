-- +goose Up
-- +goose StatementBegin

-- Add 'status' column to track the invitation status of game participants.
-- Statuses: invited, accepted, declined, confirmed
ALTER TABLE game_participants ADD COLUMN status TEXT NOT NULL DEFAULT 'invited' CHECK (status IN ('invited', 'accepted', 'declined', 'confirmed'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE game_participants DROP COLUMN status;

-- +goose StatementEnd

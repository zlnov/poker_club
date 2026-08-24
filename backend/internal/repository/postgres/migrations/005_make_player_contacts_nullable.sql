-- +goose Up

-- +goose StatementBegin

-- Make phone number and email optional.
-- Telegram does not provide the user's phone number or email address
-- in regular user/group updates, so these fields may be unavailable
-- when a Player is created automatically from Telegram data.
-- The fields remain UNIQUE, but NULL values are allowed until the
-- user provides the corresponding contact information.

ALTER TABLE players
    ALTER COLUMN phone_number DROP NOT NULL,
    ALTER COLUMN email DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

-- Restore phone number and email as required fields.
-- This rollback is only valid if there are no players with NULL
-- values in these columns.

ALTER TABLE players
    ALTER COLUMN phone_number SET NOT NULL,
    ALTER COLUMN email SET NOT NULL;

-- +goose StatementEnd
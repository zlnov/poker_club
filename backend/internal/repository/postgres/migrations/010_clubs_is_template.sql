-- +goose Up
-- +goose StatementBegin

-- Marks the single Asgard demo template club in poker_club_demo.
-- Session clones always have is_template = FALSE.
ALTER TABLE clubs
    ADD COLUMN IF NOT EXISTS is_template BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_clubs_is_template
    ON clubs (is_template)
    WHERE is_template = TRUE;

-- Monotonic counter for unique synthetic identities when cloning a template.
CREATE SEQUENCE IF NOT EXISTS demo_clone_seq START 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP SEQUENCE IF EXISTS demo_clone_seq;
DROP INDEX IF EXISTS idx_clubs_is_template;
ALTER TABLE clubs DROP COLUMN IF EXISTS is_template;

-- +goose StatementEnd

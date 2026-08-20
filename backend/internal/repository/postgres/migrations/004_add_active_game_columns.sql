-- +goose Up
-- +goose StatementBegin

-- Add timer state columns to games for Cash games with duration.
-- timer_paused_at: NULL when timer is running, set to timestamp when paused
-- timer_paused_duration: cumulative time the timer has been paused
-- timer_notified: TRUE when time expiration notification has been sent (to avoid duplicate notifications)
ALTER TABLE games ADD COLUMN timer_paused_at TIMESTAMPTZ;
ALTER TABLE games ADD COLUMN timer_paused_duration INTERVAL NOT NULL DEFAULT '00:00:00';
ALTER TABLE games ADD COLUMN timer_notified BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE games DROP COLUMN IF EXISTS timer_notified;
ALTER TABLE games DROP COLUMN IF EXISTS timer_paused_duration;
ALTER TABLE games DROP COLUMN IF EXISTS timer_paused_at;

-- +goose StatementEnd

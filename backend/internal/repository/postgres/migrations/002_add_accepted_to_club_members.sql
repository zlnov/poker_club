-- +goose Up
-- +goose StatementBegin

-- Add 'accepted' column to track whether an invited user has accepted the invitation.
-- This distinguishes "invited but not responded" (accepted=FALSE) from
-- "accepted by user but awaiting owner/admin confirmation" (accepted=TRUE).
ALTER TABLE club_members ADD COLUMN accepted BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE club_members DROP COLUMN accepted;

-- +goose StatementEnd

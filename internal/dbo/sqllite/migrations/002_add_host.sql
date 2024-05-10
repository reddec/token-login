-- +migrate Up
ALTER TABLE token
    ADD COLUMN host TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE token
    DROP COLUMN host;
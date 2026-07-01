-- +migrate Up
ALTER TABLE token ADD COLUMN IF NOT EXISTS host TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE token DROP COLUMN host;

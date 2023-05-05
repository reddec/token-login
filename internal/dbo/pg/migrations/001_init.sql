-- +migrate Up
CREATE TABLE token
(
    id             BIGSERIAL   NOT NULL PRIMARY KEY,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    key_id         TEXT        NOT NULL,
    hash           BYTEA       NOT NULL,
    "user"         TEXT        NOT NULL,
    label          TEXT        NOT NULL DEFAULT '',
    path           TEXT        NOT NULL DEFAULT '/**',
    headers        JSONB,
    requests       BIGINT      NOT NULL DEFAULT 0,
    last_access_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE UNIQUE INDEX token_key_id ON token (key_id);
CREATE INDEX token_user ON token ("user");

-- +migrate Down
DROP INDEX token_user;
DROP INDEX token_key_id;
DROP TABLE token;
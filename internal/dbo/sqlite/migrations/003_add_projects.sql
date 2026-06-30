-- +migrate Up
CREATE TABLE IF NOT EXISTS project
(
    id          INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME NOT NULL DEFAULT current_timestamp,
    updated_at  DATETIME NOT NULL DEFAULT current_timestamp,
    user        TEXT     NOT NULL DEFAULT '',
    slug        TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS project_user_slug ON project (user, slug);

-- Pre-fill default project for each user who already has tokens
INSERT INTO project ("user", slug, description)
SELECT DISTINCT "user", '', 'Default project'
FROM token
WHERE true
ON CONFLICT DO NOTHING;

-- Rebuild token table with project_id as NOT NULL (SQLite lacks ALTER COLUMN SET NOT NULL)
CREATE TABLE token_new
(
    id             INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at     DATETIME NOT NULL DEFAULT current_timestamp,
    updated_at     DATETIME NOT NULL DEFAULT current_timestamp,
    key_id         TEXT     NOT NULL UNIQUE,
    hash           BLOB     NOT NULL,
    user           TEXT     NOT NULL,
    label          TEXT     NOT NULL DEFAULT '',
    path           TEXT     NOT NULL DEFAULT '/**',
    host           TEXT     NOT NULL DEFAULT '',
    headers        JSON,
    requests       INTEGER  NOT NULL DEFAULT 0,
    last_access_at DATETIME NOT NULL DEFAULT current_timestamp,
    project_id     INTEGER  NOT NULL REFERENCES project(id) ON DELETE CASCADE
);

-- Copy existing tokens and assign each to its user's default project
INSERT INTO token_new (id, created_at, updated_at, key_id, hash, user, label, path, host, headers, requests, last_access_at, project_id)
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t.user, t.label, t.path, t.host, t.headers, t.requests, t.last_access_at, p.id
FROM token t
JOIN project p ON p.user = t.user AND p.slug = '';

DROP VIEW IF EXISTS token_view;
DROP TABLE token;

ALTER TABLE token_new RENAME TO token;

CREATE UNIQUE INDEX IF NOT EXISTS token_key_id ON token (key_id);
CREATE INDEX IF NOT EXISTS token_user ON token (user);

CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t.user, t.label, t.path, t.host, t.headers, t.requests, t.last_access_at, t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

-- +migrate Down
DROP VIEW IF EXISTS token_view;

CREATE TABLE token_old
(
    id             INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at     DATETIME NOT NULL DEFAULT current_timestamp,
    updated_at     DATETIME NOT NULL DEFAULT current_timestamp,
    key_id         TEXT     NOT NULL UNIQUE,
    hash           BLOB     NOT NULL,
    user           TEXT     NOT NULL,
    label          TEXT     NOT NULL DEFAULT '',
    path           TEXT     NOT NULL DEFAULT '/**',
    host           TEXT     NOT NULL DEFAULT '',
    headers        JSON,
    requests       INTEGER  NOT NULL DEFAULT 0,
    last_access_at DATETIME NOT NULL DEFAULT current_timestamp
);

INSERT INTO token_old (id, created_at, updated_at, key_id, hash, user, label, path, host, headers, requests, last_access_at)
SELECT id, created_at, updated_at, key_id, hash, user, label, path, host, headers, requests, last_access_at FROM token;

DROP TABLE token;

ALTER TABLE token_old RENAME TO token;

CREATE UNIQUE INDEX IF NOT EXISTS token_key_id ON token (key_id);
CREATE INDEX IF NOT EXISTS token_user ON token (user);

DROP INDEX IF EXISTS project_user_slug;
DROP TABLE IF EXISTS project;

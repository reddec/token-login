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

ALTER TABLE token ADD COLUMN project_id INTEGER REFERENCES project(id) ON DELETE CASCADE;

-- Assign all existing tokens to their user's default project
UPDATE token SET project_id = (
    SELECT id FROM project WHERE project."user" = token."user" AND project.slug = ''
);

CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t.user, t.label, t.path, t.host, t.headers, t.requests, t.last_access_at, t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

-- +migrate Down
DROP VIEW IF EXISTS token_view;
ALTER TABLE token DROP COLUMN project_id;
DROP INDEX IF EXISTS project_user_slug;
DROP TABLE IF EXISTS project;

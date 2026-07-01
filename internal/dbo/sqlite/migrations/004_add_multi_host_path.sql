-- +migrate Up
ALTER TABLE token ADD COLUMN hosts TEXT NOT NULL DEFAULT '[]';
ALTER TABLE token ADD COLUMN paths TEXT NOT NULL DEFAULT '["/**"]';

-- Migrate existing single host/path values to JSON arrays.
-- Empty host -> [], non-empty -> ["value"]. Same for path.
UPDATE token SET hosts = CASE WHEN host = '' THEN '[]' ELSE json_array(host) END;
UPDATE token SET paths = CASE WHEN path = '' OR path = '/**' THEN '["/**"]' ELSE json_array(path) END;

-- Drop view that references old columns first.
DROP VIEW IF EXISTS token_view;

-- SQLite >= 3.35 supports DROP COLUMN.
ALTER TABLE token DROP COLUMN host;
ALTER TABLE token DROP COLUMN path;

-- Recreate token_view with new column names.
CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t.user, t.label,
       t.hosts, t.paths, t.headers, t.requests, t.last_access_at,
       t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

-- +migrate Down
DROP VIEW IF EXISTS token_view;
ALTER TABLE token ADD COLUMN host TEXT NOT NULL DEFAULT '';
ALTER TABLE token ADD COLUMN path TEXT NOT NULL DEFAULT '/**';
UPDATE token SET host = COALESCE(json_extract(hosts, '$[0]'), ''),
                path = COALESCE(json_extract(paths, '$[0]'), '/**');
ALTER TABLE token DROP COLUMN hosts;
ALTER TABLE token DROP COLUMN paths;
CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t.user, t.label,
       t.host, t.path, t.headers, t.requests, t.last_access_at,
       t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

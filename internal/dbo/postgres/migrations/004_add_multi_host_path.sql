-- +migrate Up
ALTER TABLE token ADD COLUMN hosts JSONB NOT NULL DEFAULT '[]';
ALTER TABLE token ADD COLUMN paths JSONB NOT NULL DEFAULT '["/**"]';

-- Migrate existing single host/path values to JSON arrays.
UPDATE token SET hosts = CASE WHEN host = '' THEN '[]'::jsonb ELSE jsonb_build_array(host) END;
UPDATE token SET paths = CASE WHEN path = '' OR path = '/**' THEN '["/**"]'::jsonb ELSE jsonb_build_array(path) END;

DROP VIEW IF EXISTS token_view;

ALTER TABLE token DROP COLUMN host;
ALTER TABLE token DROP COLUMN path;

-- Recreate token_view.
CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t."user", t.label,
       t.hosts, t.paths, t.headers, t.requests, t.last_access_at,
       t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

-- +migrate Down
DROP VIEW IF EXISTS token_view;
ALTER TABLE token ADD COLUMN host TEXT NOT NULL DEFAULT '';
ALTER TABLE token ADD COLUMN path TEXT NOT NULL DEFAULT '/**';
UPDATE token SET host = COALESCE(hosts->>0, ''),
                path = COALESCE(paths->>0, '/**');
ALTER TABLE token DROP COLUMN hosts;
ALTER TABLE token DROP COLUMN paths;
CREATE VIEW token_view AS
SELECT t.id, t.created_at, t.updated_at, t.key_id, t.hash, t."user", t.label,
       t.host, t.path, t.headers, t.requests, t.last_access_at,
       t.project_id, p.slug AS project_slug
FROM token t
JOIN project p ON t.project_id = p.id;

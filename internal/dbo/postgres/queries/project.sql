-- name: GetProject :one
SELECT * FROM project WHERE "user" = $1 AND id = $2;

-- name: ListProjects :many
SELECT * FROM project WHERE "user" = $1 ORDER BY id ASC;

-- name: ListAllProjects :many
SELECT * FROM project;

-- name: CreateProject :one
INSERT INTO project ("user", slug, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateProject :execrows
UPDATE project SET description = $1, updated_at = now()
WHERE "user" = $2 AND id = $3;

-- name: DeleteProject :execrows
DELETE FROM project WHERE "user" = $1 AND id = $2;

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM project WHERE "user" = $1 AND id = $2) AS ok;

-- name: GetDefaultProject :one
SELECT * FROM project WHERE "user" = $1 AND slug = '';

-- name: CreateDefaultProject :one
INSERT INTO project ("user", slug, description)
VALUES ($1, '', 'Default project')
RETURNING *;

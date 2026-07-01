-- name: GetProject :one
SELECT * FROM project WHERE "user" = ? AND id = ?;

-- name: ListProjects :many
SELECT * FROM project WHERE "user" = ? ORDER BY id ASC;

-- name: ListAllProjects :many
SELECT * FROM project;

-- name: CreateProject :one
INSERT INTO project ("user", slug, description)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateProject :execrows
UPDATE project SET description = ?, updated_at = current_timestamp
WHERE "user" = ? AND id = ?;

-- name: DeleteProject :execrows
DELETE FROM project WHERE "user" = ? AND id = ?;

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM project WHERE "user" = ? AND id = ?) AS ok;


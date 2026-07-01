-- name: GetToken :one
SELECT * FROM token_view WHERE "user" = $1 AND id = $2;

-- name: GetTokenByID :one
SELECT * FROM token_view WHERE id = $1;

-- name: ListTokens :many
SELECT * FROM token_view WHERE "user" = $1 ORDER BY id DESC;

-- name: ListTokensByUserAndProject :many
SELECT * FROM token_view WHERE "user" = $1 AND project_id = $2 ORDER BY id DESC;

-- name: ListAllTokens :many
SELECT * FROM token_view;

-- name: CreateToken :one
INSERT INTO token (key_id, hash, "user", label, paths, hosts, headers, project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: UpdateToken :execrows
UPDATE token
SET hosts = $1, paths = $2, label = $3, headers = $4, updated_at = now()
WHERE "user" = $5 AND id = $6;

-- name: RefreshToken :execrows
UPDATE token
SET hash = $1, key_id = $2, updated_at = now()
WHERE "user" = $3 AND id = $4;

-- name: DeleteToken :execrows
DELETE FROM token WHERE "user" = $1 AND id = $2;

-- name: UpdateTokenStats :exec
UPDATE token
SET requests = requests + sqlc.arg(requests), last_access_at = sqlc.arg(last_access_at), updated_at = now()
WHERE id = sqlc.arg(id);

-- name: ListTokenIDsByProject :many
SELECT id FROM token WHERE project_id = $1;


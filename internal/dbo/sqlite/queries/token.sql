-- name: GetToken :one
SELECT * FROM token_view WHERE user = ? AND id = ?;

-- name: GetTokenByID :one
SELECT * FROM token_view WHERE id = ?;

-- name: ListTokens :many
SELECT * FROM token_view WHERE user = ? ORDER BY id DESC;

-- name: ListTokensByUserAndProject :many
SELECT * FROM token_view WHERE user = ? AND project_id = ? ORDER BY id DESC;

-- name: ListAllTokens :many
SELECT * FROM token_view;

-- name: CreateToken :one
INSERT INTO token (key_id, hash, user, label, paths, hosts, headers, project_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateToken :execrows
UPDATE token
SET hosts = ?, paths = ?, label = ?, headers = ?, updated_at = current_timestamp
WHERE user = ? AND id = ?;

-- name: RefreshToken :execrows
UPDATE token
SET hash = ?, key_id = ?, updated_at = current_timestamp
WHERE user = ? AND id = ?;

-- name: DeleteToken :execrows
DELETE FROM token WHERE user = ? AND id = ?;

-- name: UpdateTokenStats :exec
UPDATE token
SET requests = requests + sqlc.arg(requests), last_access_at = sqlc.arg(last_access_at), updated_at = current_timestamp
WHERE id = sqlc.arg(id);

-- name: ListTokenIDsByProject :many
SELECT id FROM token WHERE project_id = ?;


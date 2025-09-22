-- name: RevokeByRefresh :one
UPDATE refresh_tokens
SET revoked_at = $2, updated_at = $2
Where token = $1
RETURNING *;

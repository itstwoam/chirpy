-- name: GetRefreshByUserID :one
SELECT * FROM refresh_tokens
WHERE user_id = $1;

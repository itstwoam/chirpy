-- name: GetUserByToken :one
SELECT * FROM users
WHERE id = $1;

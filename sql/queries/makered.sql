-- name: MakeRed :execrows
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;

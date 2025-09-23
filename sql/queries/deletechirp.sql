-- name: DeleteChirp :execrows
DELETE FROM chirps
WHERE id = $1;

-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
Values (
  $1,
  $2
) RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

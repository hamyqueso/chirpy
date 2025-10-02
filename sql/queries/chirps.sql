-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
Values (
  $1,
  $2
) RETURNING *;

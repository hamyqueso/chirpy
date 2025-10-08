-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
  $1,
  $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: CheckUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

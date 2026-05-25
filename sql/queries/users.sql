-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
    $1,
    $2
)
RETURNING id, email, created_at, updated_at;

-- name: ListAllUsers :many
SELECT id, email, created_at, updated_at FROM users;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByID :one
SELECT id, email, created_at, updated_at FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

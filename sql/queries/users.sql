-- name: CreateUser :one
INSERT INTO users (email)
VALUES (
    $1
)
RETURNING id, email, created_at, updated_at;

-- name: ListAllUsers :many
SELECT * FROM users;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

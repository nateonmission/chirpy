-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
    $1,
    $2
)
RETURNING id, email, created_at, updated_at, is_chirpy_red;

-- name: ListAllUsers :many
SELECT id, email, created_at, updated_at, is_chirpy_red FROM users;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByID :one
SELECT id, email, created_at, updated_at, is_chirpy_red FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, email, created_at, updated_at, is_chirpy_red;

-- name: UpgradeToRed :one
UPDATE users
SET is_chirpy_red = true, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DowngradeFromRed :one
UPDATE users
SET is_chirpy_red = false, updated_at = NOW()
WHERE id = $1
RETURNING id, email, created_at, updated_at, is_chirpy_red;

-- name: IsUserChirpyRed :one
SELECT is_chirpy_red FROM users WHERE id = $1;

-- name: GetAllChirpyRedUsers :many
SELECT id, email, created_at, updated_at, is_chirpy_red FROM users WHERE is_chirpy_red = true;

-- name: GetAllNonRedUsers :many
SELECT id, email, created_at, updated_at, is_chirpy_red FROM users WHERE is_chirpy_red = false;


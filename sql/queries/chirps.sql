-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES (
    $1,
    $2
)
RETURNING id, body, created_at, updated_at, user_id;

-- name: ListAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps WHERE id = $1;

-- name: GetChirpByUser :many
SELECT * FROM chirps WHERE user_id = $1;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;
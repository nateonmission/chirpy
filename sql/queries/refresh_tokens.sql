-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: GetUserFromRefreshToken :one
SELECT users.id, users.email, users.created_at, users.updated_at
FROM users
JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
AND refresh_tokens.revoked_at IS NULL
AND refresh_tokens.expires_at > NOW();

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: DeleteAllRefreshTokens :exec
DELETE FROM refresh_tokens;
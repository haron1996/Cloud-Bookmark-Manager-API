-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_token (account_id, token_hash, token_expiry) VALUES ($1, $2, $3) ON CONFLICT (account_id) DO UPDATE SET token_hash = EXCLUDED.token_hash, token_expiry = EXCLUDED.token_expiry RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_token WHERE token_hash = $1 LIMIT 1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_token WHERE token_hash = $1;